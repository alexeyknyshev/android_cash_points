#ifndef CASHPOINTRESPONSE_H
#define CASHPOINTRESPONSE_H

#include <QtCore/QJsonObject>
#include <QtCore/QList>

struct CashPointResponse
{
    void addCashPoint(const QJsonObject &o);

    QList<QJsonObject> data;
};

#endif // CASHPOINTRESPONSE_H
