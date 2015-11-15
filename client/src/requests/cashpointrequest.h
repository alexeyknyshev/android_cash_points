#ifndef CASHPOINTREQUEST_H
#define CASHPOINTREQUEST_H

#include <QtCore/QObject>
#include <QtCore/QDateTime>

class ServerApi;
class CashPointSqlModel;

class CashPointRequest : public QObject
{
    Q_OBJECT

public:
    CashPointRequest(CashPointSqlModel *model);

    virtual void sendImpl(ServerApi *api, quint32 leftAttepmts) = 0;
    virtual void fromJson(const QJsonObject &json) = 0;

    const QDateTime &getLastUpdateTime() const { return mLastUpdateTime; }

signals:
    void update(quint32 leftAttempts);
    void error(QString err);

public slots:
    void send(quint32 leftAttempts);

protected:
    CashPointSqlModel *getModel() const { return mModel; }

    void emitUpdate(quint32 leftAttempts);
    void emitError(QString err);

    void setLastUpdateTime(const QDateTime &time);

private:
    CashPointSqlModel *const mModel;
    QDateTime mLastUpdateTime;
};


#endif // CASHPOINTREQUEST_H
