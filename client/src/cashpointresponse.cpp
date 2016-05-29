#include "cashpointresponse.h"

void CashPointResponse::addCashPointData(const QJsonObject &o)
{
    const int id = o["id"].toInt();
    if (id > 0) {
        cashPointData.insert(id, o);
    }
}

void CashPointResponse::addVisiableCashpoint(quint32 id)
{
    visiableSet.insert(id);
}

void CashPointResponse::addClusterData(const QJsonObject &o)
{
    clusterData.append(o);
}
