#ifndef NEARBYCLUSTERREQUESTFACTORY_H
#define NEARBYCLUSTERREQUESTFACTORY_H

#include "requestfactory.h"

class NearbyClusterRequestFactory : public RequestFactory
{
public:
    CashPointRequest *createRequest(CashPointSqlModel *model, const QJSValue &) const override;
    const QString &getName() const override;
};

#endif // NEARBYCLUSTERREQUESTFACTORY_H
