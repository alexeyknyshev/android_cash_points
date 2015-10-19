#ifndef TOWNLISTSQLMODEL_H
#define TOWNLISTSQLMODEL_H

#include <QtGui/QStandardItemModel>
#include <QtSql/QSqlQuery>

class ServerApi;

class TownListSqlModel : public QStandardItemModel
{
    Q_OBJECT

public:
    enum Roles {
        IdRole = Qt::UserRole,
        NameRole,
        NameTrRole,
        RegionRole,
        RolesCount = 4
    };

    explicit TownListSqlModel(QString connctionName, ServerApi *api);

    QVariant data(const QModelIndex &item, int role) const override;

public slots:
    void setFilter(QString filterStr);

    void updateFromServer();
    void updateFromServer(quint32 leftAttempts);
    void syncTowns(quint32 leftAttempts);

signals:
    void retryUpdate(quint32 leftAttempts);
    void syncNextTown(quint32 leftAttempts);

protected:
    QHash<int, QByteArray> roleNames() const override;
    int getAttemptsCount() const { return mRequestAttemptsCount; }
    int getBatchSize() const { return mRequestBatchSize; }

private slots:
    void emitRetryUpdate(quint32 leftAttempts);
    void emitSyncNextTown(quint32 leftAttempts);

private:
    ServerApi *mApi;

    QHash<int, QByteArray> mRoleNames;
    QSqlQuery mQuery;
    QSqlQuery mQueryUpdateTowns;
    QSqlQuery mQueryUpdateRegions;

    QList<int> mTownsToProcess;

    int mRequestAttemptsCount;
    int mRequestBatchSize;
};

#endif // TOWNLISTSQLMODEL_H
