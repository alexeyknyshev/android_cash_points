#ifndef CASHPOINTCREATEFACTORY_H
#define CASHPOINTCREATEFACTORY_H

#include "requestfactory.h"

class CashPointCreateFactory : public RequestFactory
{
public:
    CashPointRequest *createRequest(CashPointSqlModel *model, const QJSValue &callback) const override;
    const QString &getName() const override;
};

#endif // CASHPOINTCREATEFACTORY_H
