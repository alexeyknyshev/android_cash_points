#ifndef CASHPOINTINRADIUS_H
#define CASHPOINTINRADIUS_H

#include <QtPositioning/QGeoCoordinate>

#include "cashpointrequest.h"
#include "../serverapi_fwd.h"

class CashPointInRadius : public CashPointRequest
{
    Q_OBJECT

public:
    CashPointInRadius(CashPointSqlModel *model);

    bool fromJson(const QJsonObject &json);

private slots:
    void fetchIds(ServerApiPtr api, quint32 leftAttempts);
    void fetchCashpoints(ServerApiPtr api, quint32 leftAttempts);

private:
    QJsonObject data;

    QList<quint32> mCashpointsToProcess;
};



#endif // CASHPOINTINRADIUS_H
