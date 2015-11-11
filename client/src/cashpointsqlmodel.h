#ifndef CASHPOINTSQLMODEL_H
#define CASHPOINTSQLMODEL_H

#include <QtSql/QSqlQuery>

#include "listsqlmodel.h"

class CashPointRequest;

class CashPointSqlModel : public ListSqlModel
{
    Q_OBJECT

public:
    enum Roles {
        IdRole = Qt::UserRole,
        TypeRole,
        BankIdRole,
        TownIdRole,
        LongitudeRole,
        LatitudeRole,
        AddressRole,
        AddressCommentRole,
        MetroNameRole,
        MainOfficeRole,
        WithoutWeekendRole,
        RoundTheClockRole,
        WorksAsShopRole,
        RubRole,
        UsdRole,
        EurRole,
        CashInRole,

        RoleLast
    };

    CashPointSqlModel(const QString &connectionName,
                      ServerApi *api,
                      IcoImageProvider *imageProvider,
                      QSettings *settings);

    QVariant data(const QModelIndex &item, int role) const override;

    bool setData(const QModelIndex &index,
                 const QVariant &value,
                 int role) override;

protected:
    void updateFromServerImpl(quint32 leftAttempts) override;
    void setFilterImpl(const QString &filter) override;

    int getLastRole() const override { return RoleLast; }

    QSqlQuery &getQuery() override { return mQuery; }

private:
    QSqlQuery mQuery;

    CashPointRequest *mRequest;
};

#endif // CASHPOINTSQLMODEL_H
