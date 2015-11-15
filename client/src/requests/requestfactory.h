#ifndef REQUESTFACTORY_H
#define REQUESTFACTORY_H

class CashPointRequest;

class RequestFactory
{
public:
    RequestFactory();
    ~RequestFactory();

    virtual CashPointRequest *createRequest() const = 0;
};

#endif // REQUESTFACTORY_H
