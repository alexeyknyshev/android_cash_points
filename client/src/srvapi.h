#ifndef SRVAPI_H
#define SRVAPI_H

#include <QtCore/QObject>
#include <QtCore/QUrl>

class QNetworkAccessManager;

class SrvApi : public QObject
{
    Q_OBJECT

public:
    SrvApi(const QUrl &srvUrl);

    void setSrvUrl(const QUrl &srvUrl);
    const QUrl &getSeverUrl() const;

public slots:
    void sendRequest(QString path);

signals:
    void responseReceived(QString json);

private:
    QUrl mSrvUrl;
    QNetworkAccessManager *mNetworkMgr;
};

#endif // SRVAPI_H
