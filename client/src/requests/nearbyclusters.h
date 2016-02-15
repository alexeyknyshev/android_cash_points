#ifndef NEARBYCLUSTERS_H
#define NEARBYCLUSTERS_H

#include "cashpointrequest.h"
#include "../serverapi_fwd.h"

class NearbyClusters : public CashPointRequest
{
    Q_OBJECT

public:
    NearbyClusters(CashPointSqlModel *model);
    ~NearbyClusters();

    virtual bool fromJson(const QJsonObject &json) override;

private slots:
    void fetchClusters(ServerApiPtr api, quint32 leftAttempts);

private:
    QJsonObject data;
};

#endif // NEARBYCLUSTERS_H
