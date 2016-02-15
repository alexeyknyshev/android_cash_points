#ifndef CASHPOINTRESPONSE_H
#define CASHPOINTRESPONSE_H

#include <QtCore/QJsonObject>
#include <QtCore/QMap>
#include <QtCore/QSet>

struct CashPointResponse
{
    void addCashPointData(const QJsonObject &o);
    void addVisiableCashpoint(quint32 id);

    void addClusterData(const QJsonObject &o);

    QMap<quint32, QJsonObject> cashPointData;
    QSet<quint32> visiableSet;

    QList<QJsonObject> clusterData;

    QString message;
};

#endif // CASHPOINTRESPONSE_H
