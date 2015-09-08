#ifndef SERVERAPI_H
#define SERVERAPI_H

#include <functional>

#include <QtCore/QObject>
#include <QtCore/QUrl>

#include <QtCore/QJsonObject>

class QNetworkAccessManager;
class QNetworkReply;

class ServerApi : public QObject
{
    Q_OBJECT

public:
    ServerApi(const QString &host);

    void setHost(const QString &host);
    void setPort(int port);

    QString getHost() const;
    int getPort() const;

    qint64 uniqueRequestId() const;

    typedef std::function<void(QString data, bool ok)> Callback;

public slots:
    void sendRequest(QString path, QJsonObject data, ServerApi::Callback callback);

private slots:
    void responseReceived(QNetworkReply *);

private:
    mutable qint64 mNextUniqueId;
    QUrl mSrvUrl;
    QNetworkAccessManager *mNetworkMgr;

    std::map<qint64, ServerApi::Callback> mCallbacks;
};

#endif // SERVERAPI_H
