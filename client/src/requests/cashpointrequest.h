#ifndef CASHPOINTREQUEST_H
#define CASHPOINTREQUEST_H

#include <QtCore/QObject>
#include <QtCore/QDateTime>
#include <QtCore/QStringList>

#include "../cashpointresponse.h"

class ServerApi;
class CashPointSqlModel;

#define STEP_HANDLER(handler) #handler

class CashPointRequest : public QObject
{
    Q_OBJECT

public:
    CashPointRequest(CashPointSqlModel *model);
    ~CashPointRequest();

    virtual bool sendImpl(ServerApi *api, quint32 leftAttempts, int step);
    virtual bool fromJson(const QJsonObject &json) = 0;
    CashPointResponse *getResponse() { return mResponse; }

    const QDateTime &getLastUpdateTime() const { return mLastUpdateTime; }

signals:
    void update(quint32 leftAttempts, int step);
    void error(QString err);
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

private slots:
    void _stepFinished(ServerApi *api, int step, bool ok);

private:
    bool mHandlersRegistered;
    QStringList mStepHandlers;
    CashPointSqlModel *const mModel;
    QDateTime mLastUpdateTime;

    bool mIsRunning;
    bool mIsDisposing;
};


#endif // CASHPOINTREQUEST_H
