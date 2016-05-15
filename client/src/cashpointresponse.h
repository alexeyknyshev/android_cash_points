#ifndef CASHPOINTRESPONSE_H
#define CASHPOINTRESPONSE_H

#include <QtCore/QJsonObject>
#include <QtCore/QMap>
#include <QtCore/QSet>

class CashPointResponse
{
public:
    CashPointResponse()
        : type(CashpointData)
    { }

    enum Type {
        CashpointData,
        CreateResult,
        EditResult
    };

    void addCashPointData(const QJsonObject &o);
    void addVisiableCashpoint(quint32 id);

    void addClusterData(const QJsonObject &o);

    QMap<quint32, QJsonObject> cashPointData;
    QSet<quint32> visiableSet;

    QList<QJsonObject> clusterData;

    QString message;

    Type type;
};

#endif // CASHPOINTRESPONSE_H
