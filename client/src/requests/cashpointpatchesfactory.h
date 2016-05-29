#ifndef CASHPOINTPATCHESFACTORY_H
#define CASHPOINTPATCHESFACTORY_H

#include "requestfactory.h"

class CashPointPatchesFactory : public RequestFactory
{
public:
    CashPointRequest *createRequest(CashPointSqlModel *model, const QJSValue &callback) const override;
    const QString &getName() const override;
};

#endif // CASHPOINTPATCHESFACTORY_H
