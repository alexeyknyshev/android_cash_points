#include "cashpointresponse.h"

void CashPointResponse::addCashPoint(const QJsonObject &o)
{
    data.append(o);
}
