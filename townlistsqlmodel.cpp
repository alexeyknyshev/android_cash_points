#include "townlistsqlmodel.h"

#include <QtSql/QSqlRecord>

TownListSqlModel::TownListSqlModel(QString connectionName)
    : QSqlQueryModel(nullptr),
      mConnectionName(connectionName)
{
    mRoleNames[IdRole]     = "town_id";
    mRoleNames[NameRole]   = "town_name";
    mRoleNames[NameTrRole] = "town_name_tr";
    mRoleNames[RegionRole] = "town_region_id";

    mRoleNames[RegionRole + 1] = "index";

    setFilter("");
}

QHash<int, QByteArray> TownListSqlModel::roleNames() const
{
    return mRoleNames;
}

int TownListSqlModel::rowCount(const QModelIndex &parent) const
{
    Q_UNUSED(parent);
    return QSqlQueryModel::rowCount();
}

QVariant TownListSqlModel::data(const QModelIndex &item, int role) const
{
    if (role < Qt::UserRole)
    {
        return QSqlQueryModel::data(item, role);
    }

    if (role == RegionRole + 1)
    {
        return item.row();
    }

    QSqlRecord rec = record(item.row());
    return rec.value(role - Qt::UserRole).toString();
}

void TownListSqlModel::setFilter(QString filterStr)
{
    filterStr.replace('_', "");
    filterStr.replace('%', "");
    filterStr.replace('*', '%');
    filterStr.replace('?', '_');
    mQueryMask = "SELECT id, name, name_tr FROM towns WHERE name LIKE '%" + filterStr +
                 "%' or name_tr LIKE '%" + filterStr + "%' or "
                 "region_id IN (SELECT id FROM regions WHERE name LIKE '%" + filterStr + "%') "
                 " ORDER BY region_id, id";
    setQuery(mQueryMask, QSqlDatabase::database(mConnectionName));
}

