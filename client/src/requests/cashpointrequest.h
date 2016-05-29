#ifndef CASHPOINTREQUEST_H
#define CASHPOINTREQUEST_H

#include <QtCore/QObject>
#include <QtCore/QDateTime>
#include <QtCore/QStringList>

#include <QtQml/QJSValue>

#include "../cashpointresponse.h"

class ServerApi;
class CashPointSqlModel;

#define STEP_HANDLER(handler) #handler

#define CHECK_JSON_TYPE_STRING(val)\
if (!val.isString() && !val.isUndefined()) {\
    return false;\
}\

#define CHECK_JSON_TYPE_STRING_STRICT(val)\
if (!val.isString()) {\
    return false;\
}\

#define CHECK_JSON_TYPE_NUMBER(val)\
if (!val.isDouble() && !val.isUndefined()) {\
    return false;\
}\

#define CHECK_JSON_TYPE_NUMBER_STRICT(val)\
if (!val.isDouble()) {\
    return false;\
}\

#define CHECK_JSON_TYPE_BOOL(val)\
if (!val.isBool() && !val.isUndefined()) {\
    return false;\
}\

#define CHECK_JSON_TYPE_BOOL_STRICT(val)\
if (!val.isBool()) {\
    return false;\
}\

#define CHECK_JSON_TYPE_OBJECT(val)\
if (!val.isObject() && !val.isUndefined()) {\
    return false;\
}\

#define CHECK_JSON_TYPE_OBJECT_STRICT(val)\
if (!val.isObject()) {\
    return false;\
}\

class CashPointRequest : public QObject
{
    Q_OBJECT

public:
    CashPointRequest(CashPointSqlModel *model, QJSValue callback = QJSValue::UndefinedValue);
    ~CashPointRequest();

    virtual bool sendImpl(ServerApi *api, quint32 leftAttempts, int step);
    virtual bool fromJson(const QJsonObject &json) = 0;
    CashPointResponse *getResponse() { return mResponse; }

    const QDateTime &getLastUpdateTime() const { return mLastUpdateTime; }

    int getId() const { return mId; }

signals:
    void update(quint32 leftAttempts, int step);
    void error(CashPointRequest *request, QString msg);
    void stepFinished(ServerApi *api, int step, bool ok, QString text);
    void responseReady(CashPointRequest *request, bool finished);

public slots:
    void send(quint32 leftAttempts, int step);
    void abort();
    void dispose();

protected:
    void registerStepHandlers(const QStringList &handlers);

    CashPointSqlModel *getModel() const { return mModel; }

    void emitUpdate(quint32 leftAttempts, int step);
    void emitError(QString err);
    void emitStepFinished(ServerApi *api, int step, bool ok, QString text);
    void emitResponseReady(bool requestFinished);

    void setLastUpdateTime(const QDateTime &time);

    void addResponse(const QJsonDocument &json);

    bool isRunning() const { return mIsRunning; }
    bool isDisposing() const { return mIsDisposing; }

    const QStringList &getStepHandlers() const { return mStepHandlers; }

    CashPointResponse *mResponse;
    int mId;

    QJSValue mCallback;

private slots:
    void _stepFinished(ServerApi *api, int step, bool ok, QString msg);

private:
    bool mHandlersRegistered;
    QStringList mStepHandlers;
    CashPointSqlModel *const mModel;
    QDateTime mLastUpdateTime;

    bool mIsRunning;
    bool mIsDisposing;
};


#endif // CASHPOINTREQUEST_H
