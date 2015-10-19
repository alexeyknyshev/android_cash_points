#include "townlistsqlmodel.h"

#include <QtSql/QSqlRecord>
#include <QtSql/QSqlError>
#include <QDebug>
#include <QJsonDocument>
#include <QJsonArray>

#include "serverapi.h"
#include "rpctype.h"

#define DEFAULT_ATTEMPTS_COUNT 3
#define DEFAULT_BATCH_SIZE 128

struct Town : public RpcType<Town>
{
    QString name;
    QString nameTr;
    float longitude;
    float latitude;
    quint32 regionId;

    Town()
        : longitude(0.0f), latitude(0.0f)
    { }

    static Town fromJsonObject(const QJsonObject &obj)
    {
        Town result = RpcType<Town>::fromJsonObject(obj);

        result.name      = obj["name"].toString();
        result.nameTr    = obj["name_tr"].toString();
        result.latitude  = obj["latitude"].toDouble();
        result.longitude = obj["longitude"].toDouble();
        result.regionId  = obj["region_id"].toInt();

        return result;
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
};

/// ================================================

TownListSqlModel::TownListSqlModel(QString connectionName, ServerApi *api)
    : QStandardItemModel(0, 4, nullptr),
      mApi(api),
      mQuery(QSqlDatabase::database(connectionName)),
      mQueryUpdateTowns(QSqlDatabase::database(connectionName)),
      mQueryUpdateRegions(QSqlDatabase::database(connectionName)),
      mRequestAttemptsCount(DEFAULT_ATTEMPTS_COUNT),
      mRequestBatchSize(DEFAULT_BATCH_SIZE)
{
    mRoleNames[IdRole]     = "town_id";
    mRoleNames[NameRole]   = "town_name";
    mRoleNames[NameTrRole] = "town_name_tr";
    mRoleNames[RegionRole] = "town_region_id";

    mRoleNames[RegionRole + 1] = "index";

    if (!mQuery.prepare("SELECT id, name, name_tr FROM towns WHERE "
                        "name LIKE :filter_a or name_tr LIKE :filter_b or "
                        "region_id IN (SELECT id FROM regions WHERE name LIKE :filter_c) "
                        "ORDER BY region_id, id"))
    {
        qDebug() << "TownListSqlModel cannot prepare query:" << mQuery.lastError().databaseText();
    }

    if (!mQueryUpdateTowns.prepare("INSERT OR REPLACE INTO towns (id, name, name_tr, region_id) "
                                   "VALUES (:id, :name, :name_tr, :region_id)"))
    {
        qDebug() << "TownListSqlModel cannot prepare query:" << mQueryUpdateTowns.lastError().databaseText();
    }

    if (!mQueryUpdateRegions.prepare("INSERT OR REPLACE INTO regions (id, name) "
                                     "VALUES (:id, :name)"))
    {
        qDebug() << "TownListSqlModel cannot prepare query:" << mQueryUpdateRegions.lastError().databaseText();
    }

    connect(this, SIGNAL(retryUpdate(quint32)),
            this, SLOT(updateFromServer(quint32)), Qt::QueuedConnection);

    connect(this, SIGNAL(syncNextTown(quint32)),
            this, SLOT(syncTowns(quint32)), Qt::QueuedConnection);

    setFilter("");
}

QHash<int, QByteArray> TownListSqlModel::roleNames() const
{
    return mRoleNames;
}

QVariant TownListSqlModel::data(const QModelIndex &item, int role) const
{
    if (role == RegionRole + 1)
    {
        return item.row();
    }

    QVariant data;

    switch (role)
    {
    case IdRole:     data = QStandardItemModel::data(index(item.row(), 0), NameRole); break;
    case NameRole:   data = QStandardItemModel::data(index(item.row(), 1), NameRole); break;
    case NameTrRole: data = QStandardItemModel::data(index(item.row(), 2), NameRole); break;
    case RegionRole: data = QStandardItemModel::data(index(item.row(), 3), NameRole); break;
    }

    return data;
}


void TownListSqlModel::setFilter(QString filterStr)
{
    filterStr.replace('_', "");
    filterStr.replace('%', "");
    filterStr.replace('*', '%');
    filterStr.replace('?', '_');

    if (!filterStr.startsWith('%'))
    {
        filterStr.prepend('%');
    }

    if (!filterStr.endsWith('%'))
    {
        filterStr.append('%');
    }

    mQuery.bindValue(0, filterStr);
    mQuery.bindValue(1, filterStr);
    mQuery.bindValue(2, filterStr);

    if (!mQuery.exec()) {
        qCritical() << mQuery.lastError().databaseText();
        Q_ASSERT_X(false, "TownListSqlModel::setFilter", "sql query error");
    }

    clear();
    int row = 0;
    while (mQuery.next())
    {
        QList<QStandardItem *> items;

        for (int i = 0; i < RolesCount; ++i) {
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

    QJsonObject obj = json.object();
    QJsonValue townsVal = obj["towns"];
    if (!townsVal.isArray()) {
        qWarning() << "Json field \"towns\" is not array";
        return townIdList;
    }

    QJsonArray arr = townsVal.toArray();

    for (const QJsonValue &val : arr) {
        static const int invalidId = -1;
        const int id = val.toInt(invalidId);
        if (id > invalidId) {
            townIdList.append(id);
        }
    }

    return townIdList;
}

QList<Town> getTownList(const QJsonDocument &json)
{
    QJsonObject obj = json.object();
    return Town::fromJsonArray(obj["towns"].toArray());
}

void TownListSqlModel::updateFromServer()
{
    updateFromServer(getAttemptsCount());
}

void TownListSqlModel::updateFromServer(quint32 leftAttempts)
{
    if (leftAttempts == 0) {
        return;
    }

    /// TODO: try to use /towns endpoint
    mApi->sendRequest("/towns/list", {},
    [&](ServerApi::HttpStatusCode code, const QByteArray &data, bool timeOut) {
        if (timeOut) {
            emitRetryUpdate(leftAttempts - 1);
            return;
        }

        if (code != ServerApi::HSC_Ok) {
            qWarning() << "Server request error: " << code;
            emitRetryUpdate(leftAttempts - 1);
            return;
        }

        QJsonParseError err;
        QJsonDocument json = QJsonDocument::fromJson(data, &err);
        if (err.error != QJsonParseError::NoError) {
            qWarning() << "Server response json parse error: " << err.errorString();
            return;
        }

        mTownsToProcess = getTownsIdList(json);
        emitSyncNextTown(getAttemptsCount());
    });

    mApi->sendRequest("/regions", {},
    [&](ServerApi::HttpStatusCode code, const QByteArray &data, bool timeOut) {
        /// TODO: separate retry handler for different queries
        if (timeOut) {
            return;
        }

        if (code != ServerApi::HSC_Ok) {
            qWarning() << "Server request error: " << code;
            return;
        }

        QJsonParseError err;
        QJsonDocument json = QJsonDocument::fromJson(data, &err);
        if (err.error != QJsonParseError::NoError) {
            qWarning() << "Server response json parse error: " << err.errorString();
            return;
        }

        QJsonObject obj = json.object();
        QJsonValue regionsJson = obj["regions"];
        if (!regionsJson.isArray()) {
            qWarning() << "Server response is not json array";
            return;
        }

        QList<Region> regions = Region::fromJsonArray(regionsJson.toArray());
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

void TownListSqlModel::emitRetryUpdate(quint32 leftAttempts)
{
    if (leftAttempts > 0) {
        emit retryUpdate(leftAttempts);
    }
}

void TownListSqlModel::emitSyncNextTown(quint32 leftAttempts)
{
    if (leftAttempts > 0) {
        emit syncNextTown(leftAttempts);
    }
}

void TownListSqlModel::syncTowns(quint32 leftAttempts)
{
    if (leftAttempts == 0) {
        qDebug() << "syncTowns no retry attempt left";
        return;
    }

    if (mTownsToProcess.empty()) {
        return;
    }

    const int townsToProcess = qMin(getBatchSize(), mTownsToProcess.size());
    if (townsToProcess == 0) {
        return;
    }

    QJsonArray requestTownsBatch;
    for (int i = 0; i < townsToProcess; i++) {
        requestTownsBatch.append(mTownsToProcess.front());
        mTownsToProcess.removeFirst();
    }

    mApi->sendRequest("/towns", { QPair<QString, QJsonValue>("towns", requestTownsBatch) },
    [&](ServerApi::HttpStatusCode code, const QByteArray &data, bool timeOut) {
        if (timeOut) {
            foreach (const QJsonValue &val, requestTownsBatch) {
                const int id = val.toInt();
                if (id > 0) {
                    mTownsToProcess.append(id);
                }
            }

            emitSyncNextTown(leftAttempts - 1);
            return;
        }

        if (code != ServerApi::HSC_Ok) {
            qWarning() << "syncTowns: http status code: " << code;
            foreach (const QJsonValue &val, requestTownsBatch) {
                const int id = val.toInt();
                if (id > 0) {
                    mTownsToProcess.append(id);
                }
            }

            emitSyncNextTown(leftAttempts - 1);
            return;
        }

        QJsonParseError err;
        QJsonDocument json = QJsonDocument::fromJson(data, &err);
        if (err.error != QJsonParseError::NoError) {
            qWarning() << "syncTowns: response parse error: " << err.errorString();
            return;
        }

        QList<Town> townList = getTownList(json);
        if (!townList.isEmpty()) {
            qWarning() << "syncTowns: response is empty\n"
                       << QString::fromUtf8(data);
        }

        for (const Town &town : townList) {
            mQueryUpdateTowns.bindValue(0, town.id);
            mQueryUpdateTowns.bindValue(1, town.name);
            mQueryUpdateTowns.bindValue(2, town.nameTr);
            mQueryUpdateTowns.bindValue(3, town.regionId);

            if (!mQueryUpdateTowns.exec()) {
                qWarning() << "syncTowns: failed to update 'towns' table";
                qWarning() << "syncTowns: " << mQueryUpdateTowns.lastError().databaseText();
            }
        }

        emitSyncNextTown(getAttemptsCount());
    });
}
