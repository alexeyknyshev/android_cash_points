#include "banklistsqlmodel.h"

#include <QtSql/QSqlRecord>
#include <QtSql/QSqlError>
#include <QtCore/QDebug>

#include "rpctype.h"

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
        : licence(0),
          rating(0)
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

/// ================================================

BankListSqlModel::BankListSqlModel(QString connectionName, ServerApi *api)
    : ListSqlModel(connectionName, api),
      mQuery(QSqlDatabase::database(connectionName))
{
    setRoleName(IdRole,        "bank_id");
    setRoleName(NameRole,      "bank_name");
    setRoleName(LicenceRole,   "bank_licence");
    setRoleName(NameTrRole,    "bank_name_tr");
    setRoleName(RaitingRole,   "bank_raiting");
    setRoleName(NameTrAltRole, "bank_name_tr_alt");
    setRoleName(TelRole,       "bank_tel");

    if (!mQuery.prepare("SELECT id, name, licence, name_tr, region, name_tr_alt, tel FROM banks"
                        " WHERE"
                        "       name LIKE :name"
                        " or licence LIKE :licence"
                        " or name_tr LIKE :name_tr"
                        " or  region LIKE :town"
                        " or     tel LIKE :tel"
                        " ORDER BY raiting"))
    {
        qDebug() << "BankListSqlModel cannot prepare query:" << mQuery.lastError().databaseText();
    }
}

QHash<int, QByteArray> BankListSqlModel::roleNames() const
{
    return mRoleNames;
}

QVariant BankListSqlModel::data(const QModelIndex &item, int role) const
{
    if (role < Qt::UserRole || role >= RoleLast)
    {
        return ListSqlModel::data(item, role);
    }

    return QStandardItemModel::data(index(item.row(), role - IdRole), role);
}

void BankListSqlModel::updateFromServerImpl(quint32 leftAttempts)
{
    if (leftAttempts == 0) {
        return;
    }

    emitUpdateBanksData(leftAttempts);
}

void BankListSqlModel::setFilterImpl(const QString &filter)
{
    /*
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
    */
}

void BankListSqlModel::syncBanks(quint32 leftAttempts)
{

}

