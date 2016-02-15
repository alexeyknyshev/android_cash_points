#include "nearbyclusterrequestfactory.h"

#include "nearbyclusters.h"

CashPointRequest *NearbyClusterRequestFactory::createRequest(CashPointSqlModel *model) const
{
    return new NearbyClusters(model);
}
