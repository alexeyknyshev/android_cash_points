#include "cashpointsqlmodel.h"

#include <QtSql/QSqlRecord>

CashPointSqlModel::CashPointSqlModel(const QString &connectionName,
                                     ServerApi *api,
                                     IcoImageProvider *imageProvider,
                                     QSettings *settings)
    : ListSqlModel(connectionName, api, imageProvider, settings),
      mQuery(QSqlDatabase::database(connectionName))
{
    setRoleName(IdRole,             "cp_id");
    setRoleName(TypeRole,           "cp_type");
    setRoleName(BankIdRole,         "cp_bank_id");
    setRoleName(TownIdRole,         "cp_town_id");
    setRoleName(LongitudeRole,      "cp_coord_lon");
    setRoleName(LatitudeRole,       "cp_coord_lat");
    setRoleName(AddressRole,        "cp_address");
    setRoleName(AddressCommentRole, "cp_address_comment");
    setRoleName(MetroNameRole,      "cp_metro_name");
    setRoleName(MainOfficeRole,     "cp_main_office");
    setRoleName(WithoutWeekendRole, "cp_without_weekend");
    setRoleName(RoundTheClockRole,  "cp_round_the_clock");
    setRoleName(WorksAsShopRole,    "cp_works_as_shop");
    setRoleName(RubRole,            "cp_rub");
    setRoleName(UsdRole,            "cp_usd");
    setRoleName(EurRole,            "cp_eur");
    setRoleName(CashInRole,         "cp_cash_in");
}

QVariant CashPointSqlModel::data(const QModelIndex &item, int role) const
{

}

bool CashPointSqlModel::setData(const QModelIndex &index, const QVariant &value, int role)
{

}

void CashPointSqlModel::updateFromServerImpl(quint32 leftAttempts)
{

}

void CashPointSqlModel::setFilterImpl(const QString &filter)
{

}
