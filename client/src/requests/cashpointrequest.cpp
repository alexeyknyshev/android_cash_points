#include "cashpointrequest.h"

#include "../cashpointsqlmodel.h"

CashPointRequest::CashPointRequest(CashPointSqlModel *model)
    : mModel(model)
{
    connect(this, SIGNAL(update(quint32)), SLOT(send(quint32)), Qt::QueuedConnection);
    connect(this, SIGNAL(error(QString)), mModel, SIGNAL(requestError(QString)));
}

void CashPointRequest::send(quint32 leftAttempts)
{
    sendImpl(getModel()->getServerApi(), leftAttempts);
}

void CashPointRequest::emitUpdate(quint32 leftAttempts) {
    emit update(leftAttempts);
}

void CashPointRequest::emitError(QString err) {
    emit error(err);
}

void CashPointRequest::setLastUpdateTime(const QDateTime &time) {
    mLastUpdateTime = time;
}
