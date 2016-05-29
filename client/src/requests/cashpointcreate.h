#ifndef CASHPOINTCREATE_H
#define CASHPOINTCREATE_H

#include "cashpointrequest.h"
#include "../serverapi_fwd.h"

class CashPointCreate : public CashPointRequest
{
    Q_OBJECT

public:
    CashPointCreate(CashPointSqlModel *model,
                    QJSValue callback = QJsonValue::Undefined);

    virtual bool fromJson(const QJsonObject &json) override;

private slots:
    void createCashpoint(ServerApiPtr api, quint32 leftAttempts);

private:
    QJsonObject data;
};

#endif // CASHPOINTCREATE_H
