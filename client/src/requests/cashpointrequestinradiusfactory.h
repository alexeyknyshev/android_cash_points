#ifndef CASHPOINTREQUESTINRADIUSFACTORY_H
#define CASHPOINTREQUESTINRADIUSFACTORY_H

#include "requestfactory.h"

class CashPointRequestInRadiusFactory : public RequestFactory
{
public:
    virtual CashPointRequest *createRequest() const;
};

#endif // CASHPOINTREQUESTINRADIUSFACTORY_H
