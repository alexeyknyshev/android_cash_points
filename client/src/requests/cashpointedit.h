#ifndef CASHPOINTEDIT_H
#define CASHPOINTEDIT_H

#include "cashpointrequest.h"
#include "../serverapi_fwd.h"

class CashPointEdit : public CashPointRequest
{
    Q_OBJECT

public:
    CashPointEdit(CashPointSqlModel *model,
                  QJSValue callback = QJSValue::UndefinedValue);

    virtual bool fromJson(const QJsonObject &json) override;

private slots:
    void editCashpoint(ServerApiPtr api, quint32 leftAttempts);

private:
    QJsonObject data;
};

#endif // CASHPOINTEDIT_H
