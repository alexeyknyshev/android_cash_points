#include "cashpointcreatefactory.h"

#include "cashpointcreate.h"

CashPointRequest *CashPointCreateFactory::createRequest(CashPointSqlModel *model, const QJSValue &callback) const
{
    return new CashPointCreate(model, callback);
}

const QString &CashPointCreateFactory::getName() const
{
    static const QString name = "cashpointCreate";
    return name;
}
