#ifndef CASHPOINTSQLMODEL_H
#define CASHPOINTSQLMODEL_H

#include <QtSql/QSqlQuery>

#include "listsqlmodel.h"

class CashPointRequest;
class RequestFactory;

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

    ~CashPointSqlModel();

    QVariant data(const QModelIndex &item, int role) const override;

    bool setData(const QModelIndex &index,
                 const QVariant &value,
                 int role) override;

    void sendRequest(CashPointRequest *request);


signals:
    void delayedUpdate();

protected:
    void updateFromServerImpl(quint32 leftAttempts) override;
    void setFilterImpl(const QString &filter) override;

    int getLastRole() const override { return RoleLast; }

    QSqlQuery &getQuery() override { return mQuery; }
    bool needEscapeFilter() const override { return false; }

private:
    void setFilterJson(const QJsonObject &json);
    void setFilterFreeForm(const QString &filter);

    QSqlQuery mQuery;
    CashPointRequest *mRequest;

    QMap<QString, RequestFactory *> mRequestFactoryMap;
};

#endif // CASHPOINTSQLMODEL_H
