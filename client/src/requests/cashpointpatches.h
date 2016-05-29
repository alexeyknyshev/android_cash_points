#ifndef CASHPOINTPATCHES_H
#define CASHPOINTPATCHES_H

#include "cashpointrequest.h"
#include "../serverapi_fwd.h"

class CashPointPatches : public CashPointRequest
{
    Q_OBJECT

public:
    CashPointPatches(CashPointSqlModel *model,
                     QJSValue callback = QJSValue::UndefinedValue);

    bool fromJson(const QJsonObject &json) override;

private slots:
    void getCashpointPatches(ServerApiPtr api, quint32 leftAttempts);
};

#endif // CASHPOINTPATCHES_H
