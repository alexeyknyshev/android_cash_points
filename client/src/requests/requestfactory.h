#ifndef REQUESTFACTORY_H
#define REQUESTFACTORY_H

class CashPointRequest;
class CashPointSqlModel;
class QJSValue;
class QString;

class RequestFactory
{
public:
    RequestFactory();
    virtual ~RequestFactory();

    virtual CashPointRequest *createRequest(CashPointSqlModel *model, const QJSValue &callback) const = 0;
    virtual const QString &getName() const = 0;
};

#endif // REQUESTFACTORY_H
