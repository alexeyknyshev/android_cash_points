#include "cashpointsqlmodel.h"

#include <QtCore/QDebug>
#include <QtCore/QJsonDocument>
#include <QtCore/QJsonArray>
#include <QtCore/QSettings>

#include <QtSql/QSqlRecord>
#include <QtSql/QSqlError>

#include "rpctype.h"
#include "serverapi.h"

#include "requests/cashpointrequest.h"
#include "requests/cashpointcreatefactory.h"
#include "requests/cashpointeditfactory.h"
#include "requests/cashpointpatchesfactory.h"
#include "requests/cashpointrequestinradiusfactory.h"
#include "requests/nearbyclusterrequestfactory.h"

// CashPoint request types
#define RT_RADUS "radius"
#define RT_TOWN "town"
#define RT_CLUSTER "cluster"
#define RT_CREATE "create"
#define RT_EDIT "edit"
#define RT_PATCHES "patches"

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
    QString schedule;
    bool mainOffice;
    bool withoutWeekend;
    bool roundTheClock;
    bool worksAsShop;
    bool freeAccess;
    QList<int> currency;
    bool cashIn;
    int timestamp;
    bool approved;
    int patchCount;

    CashPoint()
        : bankId(0),
          townId(0),
          longitude(0.0f),
          latitude(0.0f),
          mainOffice(false),
          withoutWeekend(false),
          roundTheClock(false),
          worksAsShop(false),
          freeAccess(false),
          cashIn(false),
          timestamp(0),
          approved(false),
          patchCount(0)
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
        result.mainOffice     = obj["main_office"].toBool();
        result.schedule       = obj["schedule"].toString();
        result.withoutWeekend = obj["without_weekend"].toBool();
        result.roundTheClock  = obj["round_the_clock"].toBool();
        result.worksAsShop    = obj["works_as_shop"].toBool();
        result.freeAccess     = obj["free_access"].toBool();
        result.timestamp      = obj["timestamp"].toInt();
        result.approved       = obj["approved"].toBool();
        result.patchCount     = obj["patch_count"].toInt();

        QJsonArray curArr = obj["currency"].toArray();
        for (const QJsonValue &cur : curArr) {
            const int val = cur.toInt();
            if (val > 0) {
                result.currency.append(val);
            }
        }

        return result;
    }

    void fillItem(QStandardItem *item) const override
    {
        item->setData(id,             CashPointSqlModel::IdRole);
        item->setData(type,           CashPointSqlModel::TypeRole);
        item->setData(bankId,         CashPointSqlModel::BankIdRole);
        item->setData(townId,         CashPointSqlModel::TownIdRole);
        item->setData(longitude,      CashPointSqlModel::LongitudeRole);
        item->setData(latitude,       CashPointSqlModel::LatitudeRole);
        item->setData(address,        CashPointSqlModel::AddressRole);
        item->setData(addressComment, CashPointSqlModel::AddressCommentRole);
        item->setData(metroName,      CashPointSqlModel::MetroNameRole);
        item->setData(mainOffice,     CashPointSqlModel::MainOfficeRole);
        item->setData(schedule,       CashPointSqlModel::ScheduleRole);
        item->setData(withoutWeekend, CashPointSqlModel::WithoutWeekendRole);
        item->setData(roundTheClock,  CashPointSqlModel::RoundTheClockRole);
        item->setData(worksAsShop,    CashPointSqlModel::WorksAsShopRole);
        item->setData(freeAccess,     CashPointSqlModel::FreeAccess);
        item->setData(cashIn,         CashPointSqlModel::CashInRole);
        item->setData(approved,       CashPointSqlModel::ApprovedRole);
        item->setData(patchCount,     CashPointSqlModel::PatchCountRole);

        item->setData(QVariant::fromValue<QList<int>>(currency), CashPointSqlModel::CurrencyRole);

//        item->setData(timestamp,      CashPointSqlModel::);
    }
};

QList<CashPoint> getCashPointList(const QJsonDocument &json)
{
    const QJsonObject obj = json.object();
    return CashPoint::fromJsonArray(obj["cashpoints"].toArray());
}

// ==============================================================

struct CashPointCluster : public RpcType<CashPointCluster>
{
    qreal longitude;
    qreal latitude;
    quint32 size;

    CashPointCluster()
        : longitude(0.0f),
          latitude(0.0f),
          size(0)
    { }

    static CashPointCluster fromJsonObject(const QJsonObject &obj)
    {
        CashPointCluster result;
        result.longitude = obj["longitude"].toDouble();
        result.latitude = obj["latitude"].toDouble();
        result.size = obj["size"].toInt();

        return result;
    }

    void fillItem(QStandardItem *item) const override
    {
//        item->setData(id,        CashPointSqlModel::IdRole);
        item->setData("cluster", CashPointSqlModel::TypeRole);
        item->setData(longitude, CashPointSqlModel::LongitudeRole);
        item->setData(latitude,  CashPointSqlModel::LatitudeRole);
        item->setData(size,      CashPointSqlModel::SizeRole);
    }
};


/// ================================================

CashPointSqlModel::CashPointSqlModel(const QString &connectionName,
                                     ServerApi *api,
                                     IcoImageProvider *imageProvider,
                                     QSettings *settings)
    : ListSqlModel(connectionName, api, imageProvider, settings),
      mQuery(QSqlDatabase::database(connectionName)),
//      mOutOfDateQuery(QSqlDatabase::database(connectionName)),
      mQueryUpdate(QSqlDatabase::database(connectionName)),
      mQueryCashpoint(QSqlDatabase::database(connectionName)),
      mLastRequest(nullptr)
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
    setRoleName(ScheduleRole,       "cp_schedule");
    setRoleName(WithoutWeekendRole, "cp_without_weekend");
    setRoleName(RoundTheClockRole,  "cp_round_the_clock");
    setRoleName(WorksAsShopRole,    "cp_works_as_shop");
    setRoleName(FreeAccess,         "cp_free_access");
    setRoleName(CurrencyRole,       "cp_currency");
    setRoleName(CashInRole,         "cp_cash_in");
    setRoleName(ApprovedRole,       "cp_approved");
    setRoleName(SizeRole,           "cp_size");
    setRoleName(PatchCountRole,     "cp_patch_count");

    if (!mQuery.prepare("SELECT id, timestamp FROM cp")) {
        qWarning() << "CashPointSqlModel cannot prepare query:" << mQuery.lastError().databaseText();
    }

//    if (!mOutOfDateQuery.prepare("SELECT id FROM cp WHERE timestamp < :timestamp")) {
//        qWarning() << "CashPointSqlModel cannot prepare query:" << mOutOfDateQuery.lastError().databaseText();
//    }

    if (!mQueryUpdate.prepare("INSERT OR REPLACE INTO cp "
                              "(id, type, bank_id, town_id, "
                              "cord_lon, cord_lat, address, "
                              "address_comment, metro_name, main_office, "
                              "without_weekend, round_the_clock, "
                              "works_as_shop, free_access, currency, cash_in, "
                              "timestamp, approved, schedule, patch_count) "
                              "VALUES "
                              "(:id, :type, :bank_id, :town_id, "
                              ":cord_lon, :cord_lat, :address, "
                              ":address_comment, :metro_name, :main_office, "
                              ":without_weekend, :round_the_clock, "
                              ":works_as_shop, :free_access, :currency, :cash_in, "
                              ":timestamp, :approved, :schedule, :patch_count)"))
    {
        qWarning() << "CashPointSqlModel cannot prepare query:"
                   << mQueryUpdate.lastError().databaseText();
    }

    if (!mQueryCashpoint.prepare("SELECT id, type, bank_id, town_id, "
                                 "cord_lon, cord_lat, address, "
                                 "address_comment, metro_name, main_office, "
                                 "without_weekend, round_the_clock, "
                                 "works_as_shop, free_access, currency, cash_in, "
                                 "timestamp, approved, schedule, patch_count "
                                 "FROM cp WHERE id = :id"))
    {
        qWarning() << "CashPointSqlModel cannot prepare query:"
                   << mQueryCashpoint.lastError().databaseText();
    }

    mRequestFactoryMap[RT_RADUS] = new CashPointRequestInRadiusFactory;
    mRequestFactoryMap[RT_CLUSTER] = new NearbyClusterRequestFactory;
    mRequestFactoryMap[RT_CREATE] = new CashPointCreateFactory;
    mRequestFactoryMap[RT_EDIT] = new CashPointEditFactory;
    mRequestFactoryMap[RT_PATCHES] = new CashPointPatchesFactory;

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
    qWarning() << "CashPointSqlModel::setData";
    return false;
}

static QString listToJson(const QList<int> &l)
{
    QJsonArray arr;
    for (int e : l) {
        arr.append(QJsonValue(e));
    }
    return QString::fromUtf8(QJsonDocument(arr).toJson());
}

void CashPointSqlModel::addCashPoint(const QJsonObject &obj)
{
    CashPoint cp = CashPoint::fromJsonObject(obj);
    if (cp.isValid()) {
        mQueryUpdate.bindValue(":id", cp.id);
        mQueryUpdate.bindValue(":type", cp.type);
        mQueryUpdate.bindValue(":bank_id", cp.bankId);
        mQueryUpdate.bindValue(":town_id", cp.townId);
        mQueryUpdate.bindValue(":cord_lon", cp.longitude);
        mQueryUpdate.bindValue(":cord_lat", cp.latitude);
        mQueryUpdate.bindValue(":address", cp.address);
        mQueryUpdate.bindValue(":address_comment", cp.addressComment);
        mQueryUpdate.bindValue(":metro_name", cp.metroName);
        mQueryUpdate.bindValue(":main_office", cp.mainOffice);
        mQueryUpdate.bindValue(":without_weekend", cp.withoutWeekend);
        mQueryUpdate.bindValue(":round_the_clock", cp.roundTheClock);
        mQueryUpdate.bindValue(":works_as_shop", cp.worksAsShop);
        mQueryUpdate.bindValue(":free_access", cp.freeAccess);
        mQueryUpdate.bindValue(":currency", listToJson(cp.currency));
        mQueryUpdate.bindValue(":cash_in", cp.cashIn);
        mQueryUpdate.bindValue(":timestamp", cp.timestamp);
        mQueryUpdate.bindValue(":approved", cp.approved);
        mQueryUpdate.bindValue(":schedule", cp.schedule);
        mQueryUpdate.bindValue(":patch_count", cp.patchCount);

        if (!mQueryUpdate.exec()) {
            qWarning() << "addCashPoints: failed to update 'cp' table";
            qWarning() << "addCashPoints: " << mQueryUpdate.lastError().databaseText();
            return;
        }
    }
}

QJsonObject CashPointSqlModel::getCachedCashpointData(quint32 id)
{
    mQueryCashpoint.bindValue(0, id);
    if (!mQueryCashpoint.exec()) {
        qWarning() << "CashPointSqlModel: failed to fetch cached cashpoint"
                   << id << "from db";
    }
    if (mQueryCashpoint.first()) {
        const int id = mQueryCashpoint.value(0).toInt();
        if (id <= 0) {
            return QJsonObject();
        }
        const QString type =           mQueryCashpoint.value(1).toString();
        const int bankId =             mQueryCashpoint.value(2).toInt();
        const int townId =             mQueryCashpoint.value(3).toInt();
        const float longitude =        mQueryCashpoint.value(4).toFloat();
        const float latitude =         mQueryCashpoint.value(5).toFloat();
        const QString address =        mQueryCashpoint.value(6).toString();
        const QString addressComment = mQueryCashpoint.value(7).toString();
        const QString metroName =      mQueryCashpoint.value(8).toString();
        const bool mainOffice =        mQueryCashpoint.value(9).toBool();
        const bool withoutWeekEnd =    mQueryCashpoint.value(10).toBool();
        const bool roundTheClock =     mQueryCashpoint.value(11).toBool();
        const bool worksAsShop =       mQueryCashpoint.value(12).toBool();
        const bool freeAccess =        mQueryCashpoint.value(13).toBool();
        const QString currencyJson =   mQueryCashpoint.value(14).toString();
        const bool cashIn =            mQueryCashpoint.value(15).toBool();
        const int timestamp =          mQueryCashpoint.value(16).toInt();
        const bool approved =          mQueryCashpoint.value(17).toBool();

        /// TODO: schedule is json object
        const QString schedule =       mQueryCashpoint.value(20).toString();
        const int patchCount =         mQueryCashpoint.value(21).toInt();

        QJsonObject o;
        o["id"] = id;
        o["type"] = type;
        o["bank_id"] = bankId;
        o["town_id"] = townId;
        o["longitude"] = longitude;
        o["latitude"] = latitude;
        o["address"] = address;
        o["address_comment"] = addressComment;
        o["metro_name"] = metroName;
        o["main_office"] = mainOffice;
        o["without_weekend"] = withoutWeekEnd;
        o["round_the_clock"] = roundTheClock;
        o["works_as_shop"] = worksAsShop;
        o["free_access"] = freeAccess;
        o["cash_in"] = cashIn;
        o["timestamp"] = timestamp;
        o["approved"] = approved;
        ///o["schedule"] = schedule;
        o["patch_count"] = patchCount;

        QJsonParseError err;
        QJsonDocument curr = QJsonDocument::fromJson(currencyJson.toUtf8(), &err);
        if (err.error == QJsonParseError::NoError && curr.isArray()) {
            o["currency"] = curr.array();
        } else {
            o["currency"] = QJsonArray();
        }

        return o;
    }
    return QJsonObject();
}

QString CashPointSqlModel::getCashpointById(quint32 id)
{
    return QString::fromUtf8(QJsonDocument(getCachedCashpointData(id)).toJson());
}

QMap<quint32, int> CashPointSqlModel::getCachedCashpoints() const
{
    QMap<quint32, int> result;
    if (!mQuery.exec()) {
        qWarning() << "CashPointSqlModel: failed to fetch cached cashpoints from db";
        return result;
    }

    while (mQuery.next()) {
        const int id = mQuery.value(0).toInt();
        const int timestamp = mQuery.value(1).toInt();
        if (id > 0) {
            result.insert(id, timestamp);
        }
    }

    return result;
}

bool CashPointSqlModel::sendRequestJson(RequestFactory *factory, const QString &data, QJSValue &callback)
{
    QString errMsg;

    QJsonParseError err;
    const QJsonDocument json = QJsonDocument::fromJson(data.toUtf8(), &err);
    if (err.error == QJsonParseError::NoError) {
        if (json.isObject()) {
            CashPointRequest *req = factory->createRequest(this, callback);
            if (req->fromJson(json.object())) {
                sendRequest(req);
                return true;
            } else {
                delete req;
                errMsg = factory->getName() + " " + trUtf8("request could not be parsed from json");
            }
        } else {
            errMsg = factory->getName() + " " + trUtf8("request must be json object");
        }
    } else {
        errMsg = factory->getName() + " " + trUtf8("malformed json");
    }

    callback.call({ QJSValue(0), QJSValue(-1), QJSValue(false), errMsg });
    emit cashPointOperationError(factory->getName().toLower(), errMsg);

    return false;
}

void CashPointSqlModel::sendRequest(CashPointRequest *request)
{
    if (!request) {
        return;
    }

    if (mLastRequest) {
        if (mLastRequest != request) {
            mLastRequest->dispose();
        } else {
            mLastRequest->abort();
        }
    }
    mLastRequest = request;

    connect(request, SIGNAL(destroyed(QObject*)),
            this, SLOT(onRequestDeleted(QObject*)));
    connect(request, SIGNAL(error(CashPointRequest*,QString)),
            this, SLOT(onRequestErrorReceived(CashPointRequest*,QString)));
    connect(request, SIGNAL(responseReady(CashPointRequest*,bool)),
            this, SLOT(onRequestDataReceived(CashPointRequest*,bool)));

    const quint32 startStepIndex = 0;
    request->send(getAttemptsCount(), startStepIndex);
    emit delayedUpdate();
}

void CashPointSqlModel::editCashPoint(QString data, QJSValue callback)
{
    sendRequestJson(mRequestFactoryMap[RT_EDIT], data, callback);
}

void CashPointSqlModel::createCashPoint(QString data, QJSValue callback)
{
    sendRequestJson(mRequestFactoryMap[RT_CREATE], data, callback);
}

void CashPointSqlModel::getCashPointPatches(QString data, QJSValue callback)
{
    sendRequestJson(mRequestFactoryMap[RT_PATCHES], data, callback);
}

QString CashPointSqlModel::getLastGeoPos() const
{
    bool ok = false;
    QJsonObject json;
    json["longitude"] = getSettings()->value("mypos/longitude").toReal(&ok);
    if (!ok) {
        json["longitude"] = 37.6155600;
    }

    json["latitude"] = getSettings()->value("mypos/latitude").toReal(&ok);
    if (!ok) {
        json["latitude"] = 55.7522200;
    }

    json["zoom"] = getSettings()->value("mypos/zoom").toReal(&ok);
    if (!ok) {
        json["zoom"] = 13;
    }

    return QString::fromUtf8(QJsonDocument(json).toJson());
}

void CashPointSqlModel::saveLastGeoPos(QString data)
{
    QJsonObject json = QJsonDocument::fromJson(data.toUtf8()).object();

    getSettings()->setValue("mypos/longitude", json["longitude"].toDouble());
    getSettings()->setValue("mypos/latitude", json["latitude"].toDouble());
    getSettings()->setValue("mypos/zoom", json["zoom"].toDouble());

    getSettings()->sync();
}

void CashPointSqlModel::setFilterImpl(const QString &filter, const QJsonObject &options)
{
    Q_UNUSED(options);
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
        emitRequestError(0, "CashPointSqlModel::setFilterImpl: request is not a json object.");
        return;
    }

    setFilterJson(json.object());
}

QStandardItem *CashPointSqlModel::getCachedItem(quint32 id, QList<QStandardItem *> &pool)
{
    QStandardItem *item = nullptr;
    if (pool.isEmpty()) {
        item = new QStandardItem;
    } else {
        item = pool.takeFirst();
    }

    mItemsHash[id] = item;
    return item;
}

void CashPointSqlModel::onRequestErrorReceived(CashPointRequest *request, QString msg)
{
    emit requestError(request->getId(), msg);
}

void CashPointSqlModel::onRequestDataReceived(CashPointRequest *request, bool reqFinished)
{
    Q_ASSERT(request);
    CashPointResponse *response = request->getResponse();
    switch (response->type) {
    case CashPointResponse::CashpointData: {
        clear();
        mItemsHash.clear();
        int count = onCashpointDataReceived(response);
        count += onClusterDataReceived(response);
        emit objectsFetched(count);
        break;
    }
    case CashPointResponse::EditResult:
        break;
    case CashPointResponse::CreateResult:
        break;
    }

    if (reqFinished) {
        qDebug() << request->metaObject()->className() << "request deleted";
        request->deleteLater();
        emit requestFinished(request->getId(), true);
    }
}

int CashPointSqlModel::onCashpointDataReceived(CashPointResponse *response)
{
    const auto objList = response->cashPointData.values();
    for (const QJsonObject &obj : objList) {
        addCashPoint(obj);
    }

    QList<QStandardItem *> itemsToBeRemoved;

    QSet<quint32> visiableLeftSet = response->visiableSet;

    /// remove invisible items
    auto it = mItemsHash.begin();
    while (it != mItemsHash.end()) {
        const quint32 id = it.key();
        if (!response->visiableSet.contains(id)) {
            QStandardItem *item = it.value();
            itemsToBeRemoved.append(item);
            it = mItemsHash.erase(it);
            visiableLeftSet.remove(id);

            QStandardItem *r = takeItem(item->row());
            Q_ASSERT_X(r == item, "CashPointSqlModel::onRequestDataReceived", "items missmatch");
        } else {
            it++;
        }
    }

    /// add missing (new) items
    for (const quint32 visibleId : visiableLeftSet) {
        QJsonObject data;

        /// prefer received data to cached
        const auto cpDataIt = response->cashPointData.find(visibleId);
        if (cpDataIt != response->cashPointData.end()) {
            data = cpDataIt.value();
        } else {
            data = getCachedCashpointData(visibleId);
        }

        const CashPoint cp = CashPoint::fromJsonObject(data);
        if (!cp.isValid()) {
            qWarning() << "Cashpoint with id" << visibleId
                       << "is not in cache as expected!";
            continue;
        }

        /// reuse old items
        QStandardItem *item = getCachedItem(visibleId, itemsToBeRemoved);
        cp.fillItem(item);
        appendRow(item);
    }

    for (QStandardItem *item : itemsToBeRemoved) {
        qDebug() << "deleted item" << item->data(IdRole).toUInt();
        delete item;
    }

    return response->visiableSet.size();
}

int CashPointSqlModel::onClusterDataReceived(CashPointResponse *response)
{
    const int rCount = rowCount();
    for (int row = 0; row < rCount; row++) {
        QStandardItem *item_ = item(row, 0);
        if (!item_) {
            qDebug() << "invalid item";
            continue;
        }
        if (item_->data(TypeRole).toString() == "cluster") {
            QStandardItem *itemToRemove = takeItem(item_->row(), item_->column());
            delete itemToRemove;
            qDebug() << "cluster removed";
        }
    }

    for (const QJsonObject &data : response->clusterData) {
        qDebug() << data;
        QStandardItem *item = new QStandardItem;
        CashPointCluster cluster = CashPointCluster::fromJsonObject(data);
        cluster.fillItem(item);
        appendRow(item);
    }

    return response->clusterData.size();
}

void CashPointSqlModel::onRequestDeleted(QObject *request)
{
    if (request == mLastRequest) {
        mLastRequest = nullptr;
    }
}

void CashPointSqlModel::setFilterJson(const QJsonObject &json)
{
    const QString type = json["type"].toString();

    CashPointRequest *req = nullptr;
    const auto it = mRequestFactoryMap.find(type);
    if (it != mRequestFactoryMap.end()) {
        RequestFactory *factory = it.value();
        req = factory->createRequest(this, QJSValue::UndefinedValue);
        const bool ready = req->fromJson(json);
        if (!ready) {
            delete req;
            emitRequestError(0, "CashPointSqlModel::setFilterJson: cannot preapre request of type '" + type + "' "
                                "from json data: " + QString::fromUtf8(QJsonDocument(json).toJson()));
            return;
        }
    } else {
        emitRequestError(0, "CashPointSqlModel::setFilterJson: unknown req type: " + type);
        return;
    }

    sendRequest(req);
}

void CashPointSqlModel::setFilterFreeForm(const QString &filter)
{
    if (filter.isEmpty()) {
        return;
    }

    const QString filterLower = filter.toLower();
    const QStringList wordList = filterLower.split(' ', QString::SkipEmptyParts);

    CashPointRequest *req = nullptr;
    if (wordList.contains(trUtf8("банкомат"))) {
        if (wordList.size() == 1) {
            const auto it = mRequestFactoryMap.find(RT_CLUSTER);
            if (it == mRequestFactoryMap.end()) {
                return;
            }
            RequestFactory *factory = it.value();
            req = factory->createRequest(this, QJSValue::UndefinedValue);
        } else {
            if (wordList.contains(trUtf8("ближайший"))) {

            } else if (wordList.contains(trUtf8("рядом"))) {

            } else if (wordList.contains(trUtf8("круглосуточн"))) {

            } else {

            }
        }
    } else if (wordList.contains(trUtf8("банк"))) {

    } else if (wordList.contains(trUtf8("офис"))) {

    } else {
    }

    sendRequest(req);
}
