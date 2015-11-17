#ifndef REQUESTFACTORY_H
#define REQUESTFACTORY_H

class CashPointRequest;
class CashPointSqlModel;

class RequestFactory
{
public:
    RequestFactory();
    ~RequestFactory();

    virtual CashPointRequest *createRequest(CashPointSqlModel *model) const = 0;
};

#endif // REQUESTFACTORY_H
