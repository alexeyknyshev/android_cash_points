#include "townlistsqlmodel.h"

#include <QtSql/QSqlRecord>
#include <QtSql/QSqlError>
#include <QtCore/QJsonDocument>
#include <QtCore/QJsonArray>
#include <QtCore/QDebug>
#include <QtCore/QSettings>
#include <QtCore/QStandardPaths>
#include <QtCore/QFile>
#include <QtCore/QDir>

#include "serverapi.h"
#include "rpctype.h"

struct Town : public RpcType<Town>
{
    QString name;
    QString nameTr;
    float longitude;
    float latitude;
    float zoom;
    quint32 regionId;
    bool regionalCenter;
    quint32 mine;

    Town()
        : longitude(0.0f),
          latitude(0.0f)
    { }

    static Town fromJsonObject(const QJsonObject &obj)
    {
        Town result = RpcType<Town>::fromJsonObject(obj);

        result.name      = obj["name"].toString();
        result.nameTr    = obj["name_tr"].toString();
        result.latitude  = obj["latitude"].toDouble();
        result.longitude = obj["longitude"].toDouble();
        result.zoom      = obj["zoom"].toDouble();
        result.regionId  = obj["region_id"].toInt(std::numeric_limits<int>::max());

        result.regionalCenter = obj["regional_center"].toBool();
        result.mine = obj["mine"].toInt();

        return result;
    }

    QJsonObject toJsonObject() const
    {
        QJsonObject json;

        json["id"] = QJsonValue(qint64(id));
        json["name"] = name;
        json["name_tr"] = nameTr;
        json["latitude"] = latitude;
        json["longitude"] = longitude;
        json["zoom"] = zoom;
        json["region_id"] = QJsonValue(qint64(regionId));
        json["regional_center"] = regionalCenter ? 1 : 0;

        return json;
    }

    void fillItem(QStandardItem *) const override
    {
        qWarning() << "Town::fillItem should never be called";
        Q_ASSERT_X(false, "Town::fillItem", "should never be called");
    }
};

struct Region : public RpcType<Region>
{
    QString name;

    static Region fromJsonObject(const QJsonObject &obj)
    {
        Region result = RpcType<Region>::fromJsonObject(obj);

        result.name = obj["name"].toString();

        return result;
    }

    void fillItem(QStandardItem *) const override
    {
        qWarning() << "Region::fillItem should never be called";
        Q_ASSERT_X(false, "Region::fillItem", "should never be called");
    }
};

/// ================================================

TownListSqlModel::TownListSqlModel(const QString &connectionName,
                                   ServerApi *api,
                                   IcoImageProvider *imageProvider,
                                   QSettings *settings)
    : ListSqlModel(connectionName, api, imageProvider, settings),
      mQuery(QSqlDatabase::database(connectionName)),
      mQueryUpdateTowns(QSqlDatabase::database(connectionName)),
      mQueryUpdateRegions(QSqlDatabase::database(connectionName))
{
    setRowCount(11000);

    setRoleName(IdRole,     "town_id");
    setRoleName(NameRole,   "town_name");
    setRoleName(NameTrRole, "town_name_tr");
    setRoleName(RegionRole, "town_region_id");
    setRoleName(CenterRole, "town_regional_center");
    setRoleName(MineRole,   "town_is_mine");

    if (!mQuery.prepare("SELECT id, name, name_tr, mine, cord_lon, cord_lat, zoom FROM towns WHERE "
                        "       name LIKE :name"
                        " or name_tr LIKE :name_tr"
                        " or region_id IN (SELECT id FROM regions WHERE name LIKE :region_name) "
                        "ORDER BY regional_center DESC, region_id, id"))
    {
        qDebug() << "TownListSqlModel cannot prepare query:" << mQuery.lastError().databaseText();
    }

    if (!mQueryUpdateTowns.prepare("INSERT OR REPLACE INTO towns (id, name, name_tr, region_id, "
                                   "regional_center, mine, cord_lon, cord_lat, zoom) "
                                   "VALUES (:id, :name, :name_tr, :region_id, "
                                   ":regional_center, :mine, :cord_lon, :cord_lat, :zoom)"))
    {
        qDebug() << "TownListSqlModel cannot prepare query:" << mQueryUpdateTowns.lastError().databaseText();
    }

    if (!mQueryUpdateRegions.prepare("INSERT OR REPLACE INTO regions (id, name) "
                                     "VALUES (:id, :name)"))
    {
        qDebug() << "TownListSqlModel cannot prepare query:" << mQueryUpdateRegions.lastError().databaseText();
    }

    connect(this, SIGNAL(updateTownsIdsRequest(quint32)),
            this, SLOT(updateTownsIds(quint32)), Qt::QueuedConnection);

    connect(this, SIGNAL(townIdsUpdated()),
            this, SLOT(restoreTownsData()));

    connect(this, SIGNAL(updateTownsDataRequest(quint32)),
            this, SLOT(updateTownsData(quint32)), Qt::QueuedConnection);

    connect(this, SIGNAL(updateRegionsRequest(quint32)),
            this, SLOT(updateRegions(quint32)), Qt::QueuedConnection);

    setFilter("");
}

QVariant TownListSqlModel::data(const QModelIndex &item, int role) const
{
    if (role < Qt::UserRole || role >= RoleLast)
    {
        return ListSqlModel::data(item, role);
    }

    return QStandardItemModel::data(index(item.row(), role - IdRole), role);
}


void TownListSqlModel::setFilterImpl(const QString &filter)
{
    for (int i = 0; i < 2; ++i) {
        mQuery.bindValue(i, filter);
    }

    if (!mQuery.exec()) {
        qCritical() << mQuery.lastError().databaseText();
        Q_ASSERT_X(false, "TownListSqlModel::setFilter", "sql query error");
    }

    clear();
    int row = 0;
    while (mQuery.next()) {
        QList<QStandardItem *> items;

        for (int i = 0; i < RoleLast - IdRole; ++i) {
            QStandardItem *item = new QStandardItem;
            item->setData(mQuery.value(i), IdRole + i);
            items.append(item);
        }

        insertRow(row, items);
        ++row;
    }
}

static QList<int> getTownsIdList(const QJsonDocument &json)
{
    QList<int> townIdList;

    if (!json.isArray()) {
        qWarning() << "getTownsIdList: expected json array";
        return townIdList;
    }

    const QJsonArray arr = json.array();

    for (const QJsonValue &val : arr) {
        static const int invalidId = -1;
        const int id = val.toInt(invalidId);
        if (id > invalidId) {
            townIdList.append(id);
        }
    }

    return townIdList;
}

static QList<Town> getTownList(const QJsonDocument &json)
{
    return Town::fromJsonArray(json.array());
}

static QJsonDocument getTownListJson(const QList<Town> &list)
{
    QJsonArray array;
    for (const Town &town : list) {
        array.append(QJsonValue(town.toJsonObject()));
    }
    return QJsonDocument(array);
}

void TownListSqlModel::updateFromServerImpl(quint32 leftAttempts)
{
    if (leftAttempts == 0) {
        return;
    }

    emitUpdateTownIds(leftAttempts);
    emitUpdateRegions(leftAttempts);
}

void TownListSqlModel::updateTownsIds(quint32 leftAttempts)
{
    if (leftAttempts == 0) {
        qDebug() << "updateTownsIds: no retry attempt left";
        emitRequestError(trUtf8("Could not connect to server after serval attempts"));
        return;
    }

    /// Get list of towns' ids
    getServerApi()->sendRequest("/towns", {},
    [&](ServerApi::RequestStatusCode reqCode, ServerApi::HttpStatusCode httpCode, const QByteArray &data) {
        if (reqCode == ServerApi::RSC_Timeout) {
            emitUpdateTownIds(leftAttempts - 1);
            return;
        }

        if (reqCode != ServerApi::RSC_Ok) {
            emitRequestError(ServerApi::requestStatusCodeText(reqCode));
            return;
        }

        if (httpCode != ServerApi::HSC_Ok) {
            qWarning() << "Server request error: " << httpCode;
            emitUpdateTownIds(leftAttempts - 1);
            return;
        }

        QJsonParseError err;
        const QJsonDocument json = QJsonDocument::fromJson(data, &err);
        if (err.error != QJsonParseError::NoError) {
            emitRequestError("Server response json parse error: " + err.errorString());
            return;
        }

        mTownsToProcess = getTownsIdList(json);
        //qDebug() << "got towns id list:" << mTownsToProcess.size();

        emitTownIdsUpdated();
    });
}

void TownListSqlModel::restoreTownsData()
{
    restoreFromCache(mTownsToProcess);
    if (!mTownsToProcess.isEmpty()) { // fetch left towns from server
        emitUpdateTownData(getAttemptsCount());
    }
}

void TownListSqlModel::updateTownsData(quint32 leftAttempts)
{
    if (leftAttempts == 0) {
        qDebug() << "updateTownsData: no retry attempt left";
        emitRequestError(trUtf8("Could not connect to server after serval attempts"));
        return;
    }

    const quint32 townsToProcess = qMin(getRequestBatchSize(), (quint32)mTownsToProcess.size());
    if (townsToProcess == 0) {
        saveInCache();
        return;
    }

    QJsonArray requestTownsBatch;
    for (quint32 i = 0; i < townsToProcess; ++i) {
        requestTownsBatch.append(mTownsToProcess.front());
        mTownsToProcess.removeFirst();
    }

    /// Get towns data from list
    getServerApi()->sendRequest("/towns", { QPair<QString, QJsonValue>("towns", requestTownsBatch) },
    [&](ServerApi::RequestStatusCode reqCode, ServerApi::HttpStatusCode httpCode, const QByteArray &data) {
        if (reqCode == ServerApi::RSC_Timeout) {
            for (const QJsonValue &val : requestTownsBatch) {
                const int id = val.toInt();
                if (id > 0) {
                    mTownsToProcess.append(id);
                }
            }

            emitUpdateTownData(leftAttempts - 1);
            return;
        }

        if (reqCode != ServerApi::RSC_Ok) {
            emitRequestError(ServerApi::requestStatusCodeText(reqCode));
            return;
        }

        if (httpCode != ServerApi::HSC_Ok) {
            qWarning() << "updateTownsData: http status code: " << httpCode;
            for (const QJsonValue &val : requestTownsBatch) {
                const int id = val.toInt();
                if (id > 0) {
                    mTownsToProcess.append(id);
                }
            }

            emitUpdateTownData(leftAttempts - 1);
            return;
        }

        QJsonParseError err;
        const QJsonDocument json = QJsonDocument::fromJson(data, &err);
        if (err.error != QJsonParseError::NoError) {
            emitRequestError("updateTownsData: response parse error: " + err.errorString());
            return;
        }

        const QList<Town> townList = getTownList(json);
        if (townList.isEmpty()) {
            qWarning() << "updateTownsData: response is empty\n"
                       << QString::fromUtf8(data);
        }

        for (const Town &town : townList) {
            writeTownToDB(town);
        }

        emitUpdateTownData(getAttemptsCount());
    });
}

void TownListSqlModel::updateRegions(quint32 leftAttempts)
{
    if (leftAttempts == 0) {
        qDebug() << "updateRegions: no retry attempt left";
        emitRequestError(trUtf8("Could not connect to server after serval attempts"));
        return;
    }

    getServerApi()->sendRequest("/regions", {},
    [&](ServerApi::RequestStatusCode reqCode, ServerApi::HttpStatusCode httpCode, const QByteArray &data) {
        if (reqCode == ServerApi::RSC_Timeout) {
            emitUpdateRegions(leftAttempts - 1);
            return;
        }

        if (reqCode != ServerApi::RSC_Ok) {
            emitRequestError(ServerApi::requestStatusCodeText(reqCode));
            return;
        }

        if (httpCode != ServerApi::HSC_Ok) {
            qWarning() << "Server request error: " << httpCode;
            emitUpdateRegions(leftAttempts - 1);
            return;
        }

        QJsonParseError err;
        const QJsonDocument json = QJsonDocument::fromJson(data, &err);
        if (err.error != QJsonParseError::NoError) {
            qWarning() << "Server response json parse error: " << err.errorString();
            return;
        }

        const QJsonObject obj = json.object();
        const QJsonValue regionsJson = obj["regions"];
        if (!regionsJson.isArray()) {
            qWarning() << "Server response is not json array";
            return;
        }

        const QList<Region> regions = Region::fromJsonArray(regionsJson.toArray());
        for (const Region &reg : regions) {
            mQueryUpdateRegions.bindValue(0, reg.id);
            mQueryUpdateRegions.bindValue(1, reg.name);

            if (!mQueryUpdateRegions.exec()) {
                qWarning() << "syncTowns: failed to update 'regions' table";
                qWarning() << "syncTowns: " << mQueryUpdateRegions.lastError().databaseText();
            }
        }
    });
}

void TownListSqlModel::saveInCache()
{
    QSqlQuery query(QSqlDatabase::database(getDBConnectionName()));
    if (!query.exec("SELECT id, name, name_tr, region_id, "
                    "regional_center, cord_lon, cord_lat, zoom FROM towns"))
    {
        qWarning() << "Failed to save town data cache due to sql error:" << query.lastError().databaseText();
        return;
    }

    QList<Town> townList;
    while (query.next()) {
        const int id = query.value(0).toInt();
        if (id > 0) {
            Town town;

            town.id = id;
            town.name = query.value(1).toString();
            town.nameTr = query.value(2).toString();
            town.regionId = query.value(3).toInt();
            town.regionalCenter = query.value(4).toBool();
            town.longitude = query.value(5).toFloat();
            town.latitude = query.value(6).toFloat();
            town.zoom = query.value(7).toFloat();

            townList.append(town);
        }
    }

    QJsonDocument json = getTownListJson(townList);
    QByteArray rawJson = json.toJson();
    QByteArray compressedJson = qCompress(rawJson);

    qDebug() << "Raw Town data:" << rawJson.size();
    qDebug() << "Compressed data:" << compressedJson.size();

    const QString appDataPath = QStandardPaths::writableLocation(QStandardPaths::AppDataLocation);
    const QString tdataPath = QDir(appDataPath).absoluteFilePath("tdata");
    QFile dataFile(tdataPath);
    if (!dataFile.open(QIODevice::WriteOnly)) {
        qWarning("Cannot open tdata file for writing!");
        return;
    }

    dataFile.write(compressedJson);
}

void TownListSqlModel::restoreFromCache(QList<int> &townIdList)
{
    const QString tdataPath = QStandardPaths::locate(QStandardPaths::AppDataLocation, "tdata");
    if (tdataPath.isEmpty()) {
        return;
    }

    QFile dataFile(tdataPath);
    if (!dataFile.open(QIODevice::ReadOnly)) {
        qWarning("Cannot open tdata file for reading!");
        return;
    }

    QFileInfo finfo(dataFile);
    const QDateTime lastModified = finfo.lastModified();
    const qint64 daySecs = 3600 * 24;

    qDebug() << "last tdata modified time:" << lastModified;

    // cache is outdated
    if (qAbs(lastModified.secsTo(QDateTime::currentDateTime())) > daySecs) {
        return;
    }

    QByteArray compressedJson = dataFile.readAll();
    dataFile.close();

    QByteArray rawJson = qUncompress(compressedJson);

    QJsonParseError err;
    const QJsonDocument json = QJsonDocument::fromJson(rawJson, &err);
    if (err.error != QJsonParseError::NoError) {
        qWarning() << "Cannot decode tdata json!" << err.errorString();

        QFile rmFile(tdataPath);
        rmFile.open(QIODevice::Truncate);
        return;
    }

    //qDebug() << "Raw data has been read:" << rawJson;

    if (!json.isArray()) {
        qWarning() << "Json array expected in tdata file!";

        QFile rmFile(tdataPath);
        rmFile.open(QIODevice::Truncate);
        return;
    }

    QList<Town> townList = getTownList(json);
    qDebug() << "Restored towns from cache:" << townList.size();

    for (auto it = townList.begin(); it != townList.end(); it++) {
        bool removed = townIdList.removeOne((int)it->id);
        if (!removed) { // no such town in server db
            it = townList.erase(it);
        } else { // town exists in server db
            writeTownToDB(*it);
        }
    }
}

void TownListSqlModel::writeTownToDB(const Town &town)
{
    mQueryUpdateTowns.bindValue(0, town.id);
    mQueryUpdateTowns.bindValue(1, town.name);
    mQueryUpdateTowns.bindValue(2, town.nameTr);
    mQueryUpdateTowns.bindValue(3, town.regionId);
    mQueryUpdateTowns.bindValue(4, town.regionalCenter);
    mQueryUpdateTowns.bindValue(5, town.mine);
    mQueryUpdateTowns.bindValue(6, town.longitude);
    mQueryUpdateTowns.bindValue(7, town.latitude);
    mQueryUpdateTowns.bindValue(8, town.zoom);

    if (!mQueryUpdateTowns.exec()) {
        qWarning() << "updateTownsData: failed to update 'towns' table";
        qWarning() << "updateTownsData: " << mQueryUpdateTowns.lastError().databaseText();
    }
}
