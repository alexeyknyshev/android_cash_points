#ifndef CASHPOINTEDITFACTORY_H
#define CASHPOINTEDITFACTORY_H

#include "requestfactory.h"

class CashPointEditFactory : public RequestFactory
{
public:
    CashPointRequest *createRequest(CashPointSqlModel *model, const QJSValue &callback) const override;
    const QString &getName() const override;
};

#endif // CASHPOINTEDITFACTORY_H
