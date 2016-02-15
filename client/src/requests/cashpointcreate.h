#ifndef CASHPOINTCREATE_H
#define CASHPOINTCREATE_H

#include "cashpointrequest.h"
#include "../serverapi_fwd.h"

class CashPointCreate : public CashPointRequest
{
public:
    CashPointCreate(CashPointSqlModel *model);

    virtual bool fromJson(const QJsonObject &json) override;

private slots:
    void createCashpoint(ServerApiPtr api, quint32 leftAttempts);

private:
    QJsonObject data;
};

#endif // CASHPOINTCREATE_H
