#include "nearbyclusterrequestfactory.h"

#include "nearbyclusters.h"

CashPointRequest *NearbyClusterRequestFactory::createRequest(CashPointSqlModel *model, const QJSValue &) const
{
    return new NearbyClusters(model);
}

const QString &NearbyClusterRequestFactory::getName() const
{
    static const QString name = "cashpointCreate";
    return name;
}
