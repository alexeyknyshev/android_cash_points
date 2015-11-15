#ifndef CASHPOINTINRADIUS_H
#define CASHPOINTINRADIUS_H

#include <QtPositioning/QGeoCoordinate>

#include "cashpointrequest.h"

class CashPointInRadius : public CashPointRequest
{
public:
    CashPointInRadius(CashPointSqlModel *model);

    void sendImpl(ServerApi *api, quint32 leftAttempts) override;
    void fromJson(const QJsonObject &json);

    void setRadius(qreal radius);

    void setCoordinate(const QGeoCoordinate &coord);

private:
    qreal mRadius;
    QGeoCoordinate mCoord;
};



#endif // CASHPOINTINRADIUS_H
