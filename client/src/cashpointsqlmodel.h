#ifndef CASHPOINTSQLMODEL_H
#define CASHPOINTSQLMODEL_H

#include <QtSql/QSqlQuery>
#include <QtCore/QJsonObject>

#include <QtQml/QJSValue>

#include "listsqlmodel.h"

class CashPointRequest;
class CashPointResponse;
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
        ScheduleRole,
        WithoutWeekendRole,
        RoundTheClockRole,
        WorksAsShopRole,
        FreeAccess,
        CurrencyRole,
        CashInRole,
        ApprovedRole,
        PatchCountRole,

        SizeRole,

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

    void addCashPoint(const QJsonObject &obj);

    QJsonObject getCachedCashpointData(quint32 id);
    QMap<quint32, int> getCachedCashpoints() const;

    Q_INVOKABLE QString getCashpointById(quint32 id);
    Q_INVOKABLE QString getLastGeoPos() const;

public slots:
    void editCashPoint(QString data, QJSValue callback);
    void createCashPoint(QString data, QJSValue callback);
    void getCashPointPatches(QString data, QJSValue callback);

    void saveLastGeoPos(QString data);

signals:
    void delayedUpdate();
    void requestFinished(int id, bool ok);
    void cashPointOperationError(QString operation, QString errText);
    void objectsFetched(int count);

protected:
    void updateFromServerImpl(quint32 leftAttempts) override
    { Q_UNUSED(leftAttempts); }

    void setFilterImpl(const QString &filter, const QJsonObject &options) override;

    int getLastRole() const override { return RoleLast; }

    QSqlQuery *getQuery() override { return &mQuery; }
    bool needEscapeFilter() const override { return false; }

    QList<int> getSelectedIdsImpl() const override { return {}; }

private slots:
    void onRequestErrorReceived(CashPointRequest *request, QString msg);
    void onRequestDataReceived(CashPointRequest *request, bool requestFinished);
    void onRequestDeleted(QObject *request);

private:
    bool sendRequestJson(RequestFactory *factory, const QString &data, QJSValue &callback);

    void setFilterJson(const QJsonObject &json);
    void setFilterFreeForm(const QString &filter);

    int onCashpointDataReceived(CashPointResponse *response);
    int onClusterDataReceived(CashPointResponse *response);

    QStandardItem *getCachedItem(quint32 id, QList<QStandardItem *> &pool);

    QMap<quint32, QStandardItem *> mItemsHash;

    mutable QSqlQuery mQuery;
    QSqlQuery mQueryUpdate;
    QSqlQuery mQueryCashpoint;

    QMap<QString, RequestFactory *> mRequestFactoryMap;

    CashPointRequest *mLastRequest;
};

#endif // CASHPOINTSQLMODEL_H
