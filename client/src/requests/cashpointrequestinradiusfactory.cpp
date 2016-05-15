#include "cashpointrequestinradiusfactory.h"

#include "cashpointinradius.h"

CashPointRequest *CashPointRequestInRadiusFactory::createRequest(CashPointSqlModel *model, const QJSValue &) const
{
    return new CashPointInRadius(model);
}

const QString &CashPointRequestInRadiusFactory::getName() const
{
    static const QString name = "cashpointInRadius";
    return name;
}
