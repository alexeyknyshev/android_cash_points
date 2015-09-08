#include "srvapi.h"

#include <QtNetwork/QNetworkAccessManager>

SrvApi::SrvApi(const QUrl &srvUrl)
{
    setSrvUrl(srvUrl);
    mNetworkMgr = new QNetworkAccessManager(this);
}

void SrvApi::setSrvUrl(const QUrl &srvUrl)
{
    mSrvUrl = srvUrl;
}

const QUrl &SrvApi::getSeverUrl() const
{
    return mSrvUrl;
}

void SrvApi::sendRequest(QString path)
{
    Q_UNIMPLEMENTED();
}

