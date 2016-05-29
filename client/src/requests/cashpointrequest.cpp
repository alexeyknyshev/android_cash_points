#include "cashpointrequest.h"

#include <QtCore/QDebug>

#include "../serverapi.h"
#include "../cashpointsqlmodel.h"

CashPointRequest::CashPointRequest(CashPointSqlModel *model, QJSValue callback)
    : mModel(model),
      mHandlersRegistered(false),
      mResponse(nullptr),
      mIsRunning(false),
      mIsDisposing(false),
      mId(-1),
      mCallback(callback)
{
    connect(this, SIGNAL(update(quint32,int)), SLOT(send(quint32,int)), Qt::QueuedConnection);
    connect(this, SIGNAL(stepFinished(ServerApi*,int,bool,QString)),
            SLOT(_stepFinished(ServerApi*,int,bool,QString)), Qt::QueuedConnection);
}

CashPointRequest::~CashPointRequest()
{
    delete mResponse;
}

bool CashPointRequest::sendImpl(ServerApi *api, quint32 leftAttempts, int step)
{
    if (leftAttempts == 0) {
        qDebug() << metaObject()->className() << ": no retry attempt left";
        return false;
    }

    if (step >= mStepHandlers.size()) {
        qWarning() << metaObject()->className() << "::sendImpl: no such handler for step: " << step;
        Q_ASSERT_X(false, "sendImpl", "no such handler for step");
        return false;
    }

    const QString &methodName = mStepHandlers[step];
    return QMetaObject::invokeMethod(this, methodName.toLatin1().data(), Qt::QueuedConnection,
                                    Q_ARG(ServerApiPtr, ServerApiPtr(api)),
                                    Q_ARG(quint32, leftAttempts));

}

void CashPointRequest::send(quint32 leftAttempts, int step)
{
    mIsRunning = true;
    sendImpl(getModel()->getServerApi(), leftAttempts, step);
}

void CashPointRequest::abort()
{
    mIsRunning = false;
}

void CashPointRequest::dispose()
{
    blockSignals(true);
    if (mIsRunning) {
        abort();
        mIsDisposing = true;
    } else {
        deleteLater();
    }
}

void CashPointRequest::registerStepHandlers(const QStringList &handlers)
{
    if (!mHandlersRegistered) {
        mHandlersRegistered = true;
        mStepHandlers = handlers;
    } else {
        qWarning() << "Step handlers already have been registred. Handlers: " << handlers;
        Q_ASSERT_X(false, "registerStepHandlers", "step handlers already have been registred");
    }
}

void CashPointRequest::_stepFinished(ServerApi *api, int step, bool ok, QString msg)
{
    mCallback.call({ QJSValue(getId()), QJSValue(step), QJSValue(ok), QJSValue(msg) });
    if (ok && step >= 0) {
        step++;
        if (step < mStepHandlers.size()) {
            sendImpl(api, getModel()->getAttemptsCount(), step);
            return;
        }
    }
    abort();
}

void CashPointRequest::emitUpdate(quint32 leftAttempts, int step)
{
    emit update(leftAttempts, step);
}

void CashPointRequest::emitError(QString err)
{
    emit error(this, err);
}

void CashPointRequest::emitStepFinished(ServerApi *api, int step, bool ok, QString text)
{
    emit stepFinished(api, step, ok, text);
}

void CashPointRequest::emitResponseReady(bool requestFinished)
{
    emit responseReady(this, requestFinished);
}

void CashPointRequest::setLastUpdateTime(const QDateTime &time) {
    mLastUpdateTime = time;
}
