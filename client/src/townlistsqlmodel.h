#ifndef TOWNLISTSQLMODEL_H
#define TOWNLISTSQLMODEL_H

#include <QtCore/QQueue>
#include <QtGui/QStandardItemModel>
#include <QtSql/QSqlQuery>

class ServerApi;

struct Town
{
    quint32 id;
    QString name;
    QString nameTr;
    float longitude;
    float latitude;
    quint32 regionId;

    Town()
        : id(0), longitude(0), latitude(0)
    { }

    bool isValid() const { return id != 0; }
};

class TownListSqlModel : public QStandardItemModel
{
    Q_OBJECT

public:
    enum Roles {
        IdRole = Qt::UserRole,
        NameRole,
        NameTrRole,
        RegionRole
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

private slots:
    void emitRetryUpdate(quint32 leftAttempts);
    void emitSyncNextTown(quint32 leftAttempts);

private:
    ServerApi *mApi;

    QHash<int, QByteArray> mRoleNames;
    QSqlQuery mQuery;
    QSqlQuery mQueryUpdateTowns;

    QQueue<int> mTownsToProcess;

    int mRequestAttemptsCount;
};

#endif // TOWNLISTSQLMODEL_H
