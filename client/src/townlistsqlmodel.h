#ifndef TOWNLISTSQLMODEL_H
#define TOWNLISTSQLMODEL_H

#include <QtSql/QSqlQuery>

#include "listsqlmodel.h"

struct Town;

class TownListSqlModel : public ListSqlModel
{
    Q_OBJECT

    friend class SearchEngine;

public:
    enum Roles {
        IdRole = Qt::UserRole,
        NameRole,
        NameTrRole,
        RegionRole,
        CenterRole,
        MineRole,

        RoleLast
    };

    TownListSqlModel(const QString &connctionName,
                     ServerApi *api,
                     IcoImageProvider *imageProvider,
                     QSettings *settings);

    QVariant data(const QModelIndex &item, int role) const override;

    QSqlQuery *getQuery() override { return &mQuery; }

signals:
    void updateRegionsRequest(quint32 leftAttempts);

    void updateTownsIdsRequest(quint32 leftAttempts);
    void updateTownsDataRequest(quint32 leftAttempts);

    void townIdsUpdated();

protected:
    void updateFromServerImpl(quint32 leftAttempts) override;
    void setFilterImpl(const QString &filter) override;

    int getLastRole() const override { return RoleLast; }

    bool needEscapeFilter() const override { return true; }

private slots:
    void restoreTownsData();
    void updateTownsIds(quint32 leftAttempts);
    void updateTownsData(quint32 leftAttempts);

    void updateRegions(quint32 leftAttempts);

private:
    void emitUpdateTownIds(quint32 leftAttempts)
    { emit updateTownsIdsRequest(leftAttempts); }

    void emitTownIdsUpdated()
    { emit townIdsUpdated(); }

    void emitUpdateTownData(quint32 leftAttempts)
    { emit updateTownsDataRequest(leftAttempts); }

    void emitUpdateRegions(quint32 leftAttempts)
    { emit updateRegionsRequest(leftAttempts); }

    void saveInCache();
    void restoreFromCache(QList<int> &townIdList);

    void writeTownToDB(const Town &town);

    QList<int> mTownsToProcess;

    QSqlQuery mQuery;
    QSqlQuery mQueryUpdateTowns;
    QSqlQuery mQueryUpdateRegions;
};

#endif // TOWNLISTSQLMODEL_H
