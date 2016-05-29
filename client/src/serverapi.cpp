#include "serverapi.h"

#include <QtNetwork/QNetworkAccessManager>
#include <QtNetwork/QNetworkRequest>
#include <QtNetwork/QNetworkReply>

#include <QtCore/QJsonDocument>

#define DEFAULT_CALLBACK_EXPIRE_TIME 1000

/// ===============================================

ServerApiPtr::ServerApiPtr()
    : d(nullptr)
{ }

ServerApiPtr::ServerApiPtr(ServerApi *api)
    : d(api)
{ }

ServerApiPtr::ServerApiPtr(const ServerApiPtr &ptr)
    : d(ptr.d)
{ }

ServerApi *ServerApiPtr::operator ->()
{
    Q_ASSERT_X(d, "ServerApiPtr::operator ->()", "nullptr of ServerApi");
    return d;
}

ServerApi *ServerApiPtr::operator *() const
{
    Q_ASSERT_X(d, "ServerApiPtr::operator *()", "nullptr of ServerApi");
    return d;
}

/// ===============================================

ServerApi::ServerApi(const QString &host, int port, QIODevice *sslCertSource, QObject *parent)
    : QObject(parent),
      mNetworkMgr(nullptr),
      mSslConfig(nullptr)
{
    // use secure connection
    if (sslCertSource) {
        _init(host, port, QSslCertificate(sslCertSource));
    } else {
        _init(host, port, QSslCertificate());
    }
}

ServerApi::ServerApi(const QString &host, int port, const QByteArray &sslCertData, QObject *parent)
    : QObject(parent),
      mNetworkMgr(nullptr),
      mSslConfig(nullptr)
{
    _init(host, port, QSslCertificate(sslCertData));
}

void ServerApi::setHost(const QString &host)
{
    mSrvUrl.setHost(host);
}

QString ServerApi::getHost() const
{
    return mSrvUrl.host();
}

void ServerApi::setPort(int port)
{
    mSrvUrl.setPort(port);
}

int ServerApi::getPort() const
{
    return mSrvUrl.port();
}

int ServerApi::uniqueRequestId() const
{
    if (std::numeric_limits<int>::max() != mNextUniqueId) {
        mNextUniqueId++;
    } else {
        mNextUniqueId = 1;
    }
    return mNextUniqueId;
}

void ServerApi::setCallbacksExpireTime(quint32 msec)
{
    mCallbacksExpiteTime = msec;
}

quint32 ServerApi::getCallbacksExpireTime() const
{
    return mCallbacksExpiteTime;
}

void ServerApi::postRequest(QString path, QJsonObject data, ServerApi::Callback callback)
{
    /// TODO: implement
    Q_ASSERT_X(false, "ServerApi::postRequest", "not implemented yet");
}

int ServerApi::sendRequest(QString path, QJsonObject data, ServerApi::Callback callback, bool auth)
{
    /// TODO: added authenticated request sending
    Q_UNUSED(auth)

    int requestId = uniqueRequestId();
    mCallbacks.insert(requestId, ExpCallback(QDateTime::currentDateTime(), callback));

    QUrl requestUrl(mSrvUrl);
    requestUrl.setPath(path);

    QNetworkRequest req;
    req.setUrl(requestUrl);
    req.setHeader(QNetworkRequest::ContentTypeHeader, "application/json; charset=utf-8");
    req.setRawHeader(QByteArray("Id"), QString::number(requestId).toUtf8());

    if (mSslConfig) {
        req.setSslConfiguration(*mSslConfig);
    }

    if (data.isEmpty()) {
        mNetworkMgr->get(req);
    } else {
        QByteArray byteData = QJsonDocument(data).toJson(QJsonDocument::Compact);
        mNetworkMgr->post(req, byteData);
#ifndef NDEBUG
        qDebug() << requestUrl.toString() << " => " << QString::fromUtf8(byteData);
#endif
    }

    return requestId;
}

void ServerApi::ping()
{
    qDebug() << "ping";
    sendRequest("/ping", {},
    [&](RequestStatusCode reqCode, HttpStatusCode httpCode, const QByteArray &data) {
        if (reqCode != RSC_Ok) {
            emitPong(false);
            return;
        }

        if (httpCode != HSC_Ok) {
            emitPong(false);
            return;
        }

        QJsonParseError err;
        QJsonDocument json = QJsonDocument::fromJson(data, &err);
        if (!json.isObject()) {
            emitPong(false);
            return;
        }

        if (json.object()["text"].toString() != "pong") {
            emitPong(false);
            return;
        }

        emitPong(true);
    });
}

void ServerApi::emitPong(bool ok)
{
    emit pong(ok);
}

static ServerApi::HttpStatusCode getHttpStatusCode(QNetworkReply *rep)
{
    bool convertOk = false;
    int status = rep->attribute(QNetworkRequest::HttpStatusCodeAttribute).toInt(&convertOk);
    if (convertOk) {
        return (ServerApi::HttpStatusCode)status;
    }
    return ServerApi::HSC_ServerError;
}

static ServerApi::RequestStatusCode getRequestStatusCode(const QNetworkReply::NetworkError err)
{
    if (err == QNetworkReply::NoError) {
        return ServerApi::RSC_Ok;
    } else if (err == QNetworkReply::ConnectionRefusedError) {
        return ServerApi::RSC_ConnectionRefused;
    } else if (err == QNetworkReply::HostNotFoundError) {
        return ServerApi::RSC_HostNotFound;
    }
    return ServerApi::RSC_Unknown;
}

static QString getRequestStatusCodeText(const ServerApi::RequestStatusCode code)
{
    QString msg = "";
    if (code == ServerApi::RSC_ConnectionRefused) {
        msg = QObject::trUtf8("Connection refused by server. ");
    } else if (code == ServerApi::RSC_HostNotFound) {
        msg = QObject::trUtf8("Sorry, server is unavaliable now. ");
    } else if (code == ServerApi::RSC_Unknown) {
        msg = QObject::trUtf8("Unknown server connection error. ");
    }

    if (!msg.isEmpty()) {
        msg += QObject::trUtf8("Please, check your internet connection and "
                               "make sure that you use recent application version.");
    }
    return msg;
}

void ServerApi::onResponseReceived(QNetworkReply *rep)
{
    QList<qint64> expiredCallbacks = _getExpiredCallbacks();

    RequestStatusCode code = getRequestStatusCode(rep->error());
    QString msg = getRequestStatusCodeText(code);

    qint64 requestId = 0;
    if (code == RSC_Ok) {
        requestId = rep->rawHeader(QByteArray("Id")).toLongLong();
    } else {
        requestId = rep->request().rawHeader(QByteArray("Id")).toLongLong();
    }

    if (requestId == 0) {
        qWarning() << "Error: cannot find out request id from request & response";
        return;
    }

    auto it = mCallbacks.find(requestId);
    if (it == mCallbacks.cend()) {
        qWarning() << "Unknown request id: " << requestId;
        return;
    }

    if (expiredCallbacks.contains(requestId)) {
        code = RSC_Timeout;
        msg = trUtf8("Response timeout exceeded.");
    }

    Callback &callback = it->callback;
    if (code == RSC_Ok) {
        callback(code, getHttpStatusCode(rep), rep->readAll());
    } else if (code == RSC_Timeout) {
        callback(code, HSC_Ok, msg.toUtf8());
    } else {
        callback(code, HSC_Ok, QByteArray());
    }
    mCallbacks.erase(it);

    _eraseExpiredCallbacks(expiredCallbacks);
}

static bool isCallbackExpired(const QDateTime &birthTime, qint64 expireTime)
{
#ifdef CP_DEBUG
    Q_UNUSED(birthTime);
    Q_UNUSED(expireTime);
    return false;
#else // CP_DEBUG
    const QDateTime now = QDateTime::currentDateTime();
    return qAbs(birthTime.msecsTo(now)) > expireTime;
#endif // CP_DEBUG
}

QList<qint64> ServerApi::_getExpiredCallbacks() const
{
    QList<qint64> erasedCallbacks;
    for (auto it = mCallbacks.begin(); it != mCallbacks.end(); it++) {
        if (isCallbackExpired(it->dt, static_cast<qint64>(getCallbacksExpireTime()))) {
            const qint64 requestId = it.key();
            erasedCallbacks.append(requestId);

            const Callback &callback = it->callback;
            callback(RSC_Timeout, HSC_Ok, QByteArray());
        }
    }
    return erasedCallbacks;
}

int ServerApi::_eraseExpiredCallbacks(const QList<qint64> &reqIdList)
{
    int count = 0;
    for (const quint64 requestId : reqIdList) {
        count += mCallbacks.remove(requestId);
    }
    return count;
}

void ServerApi::_init(const QString &host, int port, const QSslCertificate &cert)
{
    mNextUniqueId = 0;

    if (!cert.isNull()) {
        mSslConfig = new QSslConfiguration;
        mSslConfig->setCaCertificates({ cert });
        mSslConfig->setProtocol(QSsl::TlsV1_2);
    }

    setCallbacksExpireTime(DEFAULT_CALLBACK_EXPIRE_TIME);
    setHost(host);
    setPort(port);
    mNetworkMgr = new QNetworkAccessManager(this);
    connect(mNetworkMgr, SIGNAL(finished(QNetworkReply*)), this, SLOT(onResponseReceived(QNetworkReply*)));

    if (cert.isNull()) {
        mSrvUrl.setScheme("http");
    } else {
        mSrvUrl.setScheme("https");
    }
}
