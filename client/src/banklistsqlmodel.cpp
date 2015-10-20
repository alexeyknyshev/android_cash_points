#include "banklistsqlmodel.h"

#include <QtSql/QSqlRecord>

#include "rpctype.h"

#define DEFAULT_ATTEMPTS_COUNT 3
#define DEFAULT_BATCH_SIZE 128

struct Bank : public RpcType<Bank>
{
    QString name;
    QString nameTr;
    QString nameTrAlt;
    QString town;
    QString tel;
    quint32 licence;
    quint32 rating;

    Bank()
        : licence(0), rating(0)
    { }

    static Bank fromJsonObject(const QJsonObject &obj)
    {
        Bank result = RpcType<Bank>::fromJsonObject(obj);

        result.name      = obj["name"].toString();
        result.nameTr    = obj["name_tr"].toString();
        result.nameTrAlt = obj["name_tr_alt"].toString();
        result.town      = obj["town"].toString();
        result.tel       = obj["tel"].toString();
        result.licence   = obj["licence"].toInt();
        result.rating    = obj["rating"].toInt();

        return result;
    }
};

BankListSqlModel::BankListSqlModel(QString connectionName)
    : QSqlQueryModel(nullptr),
      mConnectionName(connectionName)
{
    mRoleNames[IdRole]        = "bank_id";
    mRoleNames[NameRole]      = "bank_name";
    mRoleNames[LicenceRole]   = "bank_licence";
    mRoleNames[NameTrRole]    = "bank_name_tr";
    mRoleNames[RaitingRole]   = "bank_raiting";
    mRoleNames[NameTrAltRole] = "bank_name_tr_alt";
    mRoleNames[TelRole]       = "bank_tel";

    mRoleNames[TelRole + 1]   = "index";

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

    if (role == TelRole + 1)
    {
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
    mQueryMask = "SELECT id, name, licence, name_tr, region, name_tr_alt, tel FROM banks"
                 " WHERE"
                 "       name LIKE '%" + filterStr + "%'"
                 " or licence LIKE '%" + filterStr + "%'"
                 " or name_tr LIKE '%" + filterStr + "%'"
                 " or  region LIKE '%" + filterStr + "%'"
                 " or     tel LIKE '%" + filterStr + "%'"
                 " ORDER BY raiting"
            ;
    setQuery(mQueryMask, QSqlDatabase::database(mConnectionName));
}

