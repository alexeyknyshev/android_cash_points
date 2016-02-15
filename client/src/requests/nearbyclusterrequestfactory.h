#ifndef NEARBYCLUSTERREQUESTFACTORY_H
#define NEARBYCLUSTERREQUESTFACTORY_H

#include "requestfactory.h"

class NearbyClusterRequestFactory : public RequestFactory
{
public:
    CashPointRequest *createRequest(CashPointSqlModel *model) const override;
};

#endif // NEARBYCLUSTERREQUESTFACTORY_H
