#include "serverapi.h"

#include <QtNetwork/QNetworkAccessManager>
#include <QtNetwork/QNetworkRequest>
#include <QtNetwork/QNetworkReply>

#include <QtCore/QJsonDocument>

ServerApi::ServerApi(const QString &host)
    : mNextUniqueId(0)
{
    mSrvUrl.setScheme("https");
    setHost(host);
    mNetworkMgr = new QNetworkAccessManager(this);
    connect(mNetworkMgr, SIGNAL(finished(QNetworkReply*)), this, SLOT(responseReceived(QNetworkReply*)));
}

void ServerApi::setHost(const QString &host)
{
    mSrvUrl.setHost(host);
}

QString ServerApi::getHost() const
{
    return mSrvUrl.host();
}

qint64 ServerApi::uniqueRequestId() const
{
    if (std::numeric_limits<qint64>::max() != mNextUniqueId) {
        return mNextUniqueId++;
    } else {
        mNextUniqueId = 0;
        return mNextUniqueId;
    }
}

void ServerApi::sendRequest(QString path, QJsonObject data,
                            std::function<void (QString, bool)> callback)
{
    qint64 requestId = uniqueRequestId();
    mCallbacks.insert({ requestId, callback });

    QUrl requestUrl(mSrvUrl);
    requestUrl.setPath(path);

    QNetworkRequest req;
    req.setUrl(requestUrl);
    req.setHeader(QNetworkRequest::ContentTypeHeader, "application/json; charset=utf-8");
    data.insert("id", requestId);

    if (data.isEmpty()) {
        mNetworkMgr->get(req);
    } else {
        mNetworkMgr->post(req, QJsonDocument(data).toJson(QJsonDocument::Compact));
    }
}

void ServerApi::responseReceived(QNetworkReply *rep)
{
    if (rep->error() == QNetworkReply::NoError) {

    } else {

    }
}

