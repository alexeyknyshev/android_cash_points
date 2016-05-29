#include "cashpointpatches.h"

CashPointPatches::CashPointPatches(CashPointSqlModel *model, QJSValue callback)
    : CashPointRequest(model, callback)
{
    registerStepHandlers({
                             STEP_HANDLER(getCashpointPatches)
                         });
}

void CashPointPatches::getCashpointPatches(ServerApiPtr api, quint32 leftAttempts)
{
    Q_ASSERT_X(false, "CashPointPatches::getCashpointPatches", "Not implemented");
}

bool CashPointPatches::fromJson(const QJsonObject &json)
{
    return false;
}

