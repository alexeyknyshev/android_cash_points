#include "banklistsqlmodel.h"

#include <QtSql/QSqlRecord>

BankListSqlModel::BankListSqlModel(QString connectionName)
    : QSqlQueryModel(nullptr),
      mConnectionName(connectionName)
{
    mRoleNames[IdRole]       = "bank_id";
    mRoleNames[NameRole]     = "bank_name";
    mRoleNames[UrlRole]      = "bank_url";
    mRoleNames[TelRole]      = "bank_tel";
    mRoleNames[TelDescrRole] = "bank_tel_description";

    mRoleNames[TelDescrRole + 1] = "index";

    setFilter("");
}

QHash<int, QByteArray> BankListSqlModel::roleNames() const
{
    return mRoleNames;
}

int BankListSqlModel::rowCount(const QModelIndex &parent) const
{
    Q_UNUSED(parent);
    return QSqlQueryModel::rowCount();
}

QVariant BankListSqlModel::data(const QModelIndex &item, int role) const
{
    if (role < Qt::UserRole)
    {
        return QSqlQueryModel::data(item, role);
    }

    if (role == TelDescrRole + 1) {
        return item.row();
    }

    QSqlRecord rec = record(item.row());
    return rec.value(role - Qt::UserRole).toString();
}

void BankListSqlModel::setFilter(QString filterStr)
{
    filterStr.replace('_', "");
    filterStr.replace('%', "");
    filterStr.replace('*', '%');
    filterStr.replace('?', '_');
    mQueryMask = "SELECT id, name, url, tel, tel_description FROM banks WHERE name LIKE '%" + filterStr +
                 "%' or url LIKE '%" + filterStr + "%' or tel LIKE '%" + filterStr + "%'";
    setQuery(mQueryMask, QSqlDatabase::database(mConnectionName));
}

