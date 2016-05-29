#include "cashpointeditfactory.h"

#include "cashpointedit.h"

CashPointRequest *CashPointEditFactory::createRequest(CashPointSqlModel *model, const QJSValue &callback) const
{
    return new CashPointEdit(model, callback);
}

const QString &CashPointEditFactory::getName() const
{
    static const QString name = "cashpointEdit";
    return name;
}
