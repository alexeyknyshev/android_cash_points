#ifndef CASHPOINTREQUESTINRADIUSFACTORY_H
#define CASHPOINTREQUESTINRADIUSFACTORY_H

#include "requestfactory.h"

class CashPointRequestInRadiusFactory : public RequestFactory
{
public:
    CashPointRequest *createRequest(CashPointSqlModel *model, const QJSValue &) const override;
    const QString &getName() const override;
};

#endif // CASHPOINTREQUESTINRADIUSFACTORY_H
