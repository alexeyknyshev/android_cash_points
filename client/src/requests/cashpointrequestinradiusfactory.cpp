#include "cashpointrequestinradiusfactory.h"

#include "cashpointinradius.h"

CashPointRequest *CashPointRequestInRadiusFactory::createRequest(CashPointSqlModel *model) const
{
    return new CashPointInRadius(model);
}

