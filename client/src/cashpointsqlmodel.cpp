#include "cashpointsqlmodel.h"

#include <QtCore/QDebug>
#include <QtSql/QSqlRecord>
#include <QtCore/QJsonDocument>
#include <QtCore/QJsonArray>

#include "rpctype.h"
#include "serverapi.h"
#include "requests/cashpointrequest.h"
#include "requests/cashpointrequestinradiusfactory.h"

// CashPoint request types
#define RT_RADUS "radius"
#define RT_TOWN "town"

struct CashPoint : public RpcType<CashPoint>
{
    QString type;
    quint32 bankId;
    quint32 townId;
    qreal longitude;
    qreal latitude;
    QString address;
    QString addressComment;
    QString metroName;
    bool mainOffice;
    bool withoutWeekend;
    bool roundTheClock;
    bool worksAsShop;
    bool rub;
    bool usd;
    bool eur;
    bool cashIn;

    CashPoint()
        : bankId(0),
          townId(0),
          longitude(0.0f),
          latitude(0.0f),
          mainOffice(false),
          withoutWeekend(false),
          roundTheClock(false),
          worksAsShop(false),
          rub(false),
          usd(false),
          eur(false),
          cashIn(false)
    { }

    static CashPoint fromJsonObject(const QJsonObject &obj)
    {
        CashPoint result = RpcType<CashPoint>::fromJsonObject(obj);

        result.type           = obj["type"].toString();
        result.bankId         = obj["bank_id"].toInt();
        result.townId         = obj["town_id"].toInt();
        result.longitude      = obj["longitude"].toDouble();
        result.latitude       = obj["latitude"].toDouble();
        result.address        = obj["address"].toString();
        result.addressComment = obj["address_comment"].toString();
        result.metroName      = obj["metro_name"].toString();
        result.mainOffice     = obj["main_office"].isBool();
        result.withoutWeekend = obj["without_weekend"].toBool();
        result.roundTheClock  = obj["round_the_clock"].toBool();
        result.worksAsShop    = obj["works_as_shop"].toBool();
        result.rub            = obj["rub"].toBool();
        result.usd            = obj["usd"].toBool();
        result.eur            = obj["eur"].toBool();

        return result;
    }
};

/// ================================================

CashPointSqlModel::CashPointSqlModel(const QString &connectionName,
                                     ServerApi *api,
                                     IcoImageProvider *imageProvider,
                                     QSettings *settings)
    : ListSqlModel(connectionName, api, imageProvider, settings),
      mQuery(QSqlDatabase::database(connectionName)),
      mRequest(nullptr)
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

    mRequestFactoryMap[RT_RADUS] = new CashPointRequestInRadiusFactory;

    connect(this, SIGNAL(delayedUpdate()), SLOT(updateFromServer()), Qt::QueuedConnection);
}

CashPointSqlModel::~CashPointSqlModel()
{
    const auto end = mRequestFactoryMap.end();
    for (auto it = mRequestFactoryMap.begin(); it != end; it++) {
        delete it.value();
    }
}

QVariant CashPointSqlModel::data(const QModelIndex &item, int role) const
{
    if (role < Qt::UserRole || role >= RoleLast)
    {
        return ListSqlModel::data(item, role);
    }
    return QStandardItemModel::data(index(item.row(), 0), role);
}

bool CashPointSqlModel::setData(const QModelIndex &index, const QVariant &value, int role)
{

}

void CashPointSqlModel::sendRequest(CashPointRequest *request)
{
    if (mRequest == request) {
        return;
    }

    if (mRequest) {
        mRequest->deleteLater();
    }

    mRequest = request;
    if (request) {
        request->send(getAttemptsCount());
    }

    emit delayedUpdate();
}

void CashPointSqlModel::updateFromServerImpl(quint32 leftAttempts)
{
    if (leftAttempts == 0) {
        return;
    }

    if (!mRequest) {
        return;
    }

    mRequest->send(leftAttempts);
}

void CashPointSqlModel::setFilterImpl(const QString &filter)
{
    if (filter.isEmpty()) {
        return;
    }

    QJsonParseError err;
    const QJsonDocument json = QJsonDocument::fromJson(filter.toUtf8(), &err);
    if (err.error != QJsonParseError::NoError) {
        setFilterFreeForm(filter);
        return;
    }

    if (!json.isObject()) {
        emitRequestError("CashPointSqlModel::setFilterImpl: Cannot local request is not a json object.");
        return;
    }

    setFilterJson(json.object());
}

void CashPointSqlModel::setFilterJson(const QJsonObject &json)
{
    const QString type = json["type"].toString();

    CashPointRequest *req = nullptr;
    const auto it = mRequestFactoryMap.find(type);
    if (it != mRequestFactoryMap.end()) {
        RequestFactory *factory = it.value();
        req = factory->createRequest();
        req->fromJson(json);
    } else {
        emitRequestError("CashPointSqlModel::setFilterJson: unknown req type: " + type);
        return;
    }

    sendRequest(req);
}

void CashPointSqlModel::setFilterFreeForm(const QString &filter)
{

}
