#include "cashpointpatchesfactory.h"

#include "cashpointpatches.h"

CashPointRequest *CashPointPatchesFactory::createRequest(CashPointSqlModel *model, const QJSValue &callback) const
{
    return new CashPointPatches(model, callback);
}

const QString &CashPointPatchesFactory::getName() const
{
    static const QString name = "cashpointPatches";
    return name;
}
