#include "banklistsqlmodel.h"
#include "icoimageprovider.h"

#include <QtSql/QSqlRecord>
#include <QtSql/QSqlError>
#include <QtCore/QStandardPaths>
#include <QtCore/QDir>
#include <QtCore/QJsonDocument>
#include <QtCore/QJsonArray>
#include <QtCore/QFile>
#include <QtCore/QDebug>
#include <QtCore/QSettings>

#include "rpctype.h"
#include "serverapi.h"

struct Bank : public RpcType<Bank>
{
    QString name;
    QString nameTr;
    QString nameTrAlt;
    QString town;
    QString tel;
    quint32 licence;
    quint32 rating;
    quint32 mine;
    QList<int> partners;

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
        result.mine      = obj["mine"].toInt();

        QJsonArray partners = obj["partners"].toArray();
        const auto partnersEnd = partners.constEnd();
        for (auto it = partners.constBegin(); it != partnersEnd; it++) {
            const int id = it->toInt();
            if (id > 0) {
                result.partners.append(id);
            }
        }

        return result;
    }

    QJsonObject toJsonObject() const
    {
        QJsonObject json;

        QJsonArray partnersArray;
        for (quint32 p : partners) {
            partnersArray.append(qint64(p));
        }

        json["id"] = QJsonValue(qint64(id));
        json["name"] = name;
        json["name_tr"] = nameTr;
        json["name_tr_alt"] = nameTrAlt;
        json["town"] = town;
        json["tel"] = tel;
        json["licence"] = qint64(licence);
        json["rating"] = qint64(rating);
        json["partners"] = partnersArray;

        return json;
    }

    void fillItem(QStandardItem *item) const override
    {
        item->setData(id,        BankListSqlModel::IdRole);
        item->setData(name,      BankListSqlModel::NameRole);
        item->setData(nameTr,    BankListSqlModel::NameTrRole);
        item->setData(nameTrAlt, BankListSqlModel::NameTrAltRole);
//        item->setData(town,      BankListSqlModel::);
        item->setData(tel,       BankListSqlModel::TelRole);
        item->setData(licence,   BankListSqlModel::LicenceRole);
        item->setData(rating,    BankListSqlModel::RatingRole);
        item->setData(mine,      BankListSqlModel::MineRole);
    }
};

/// ================================================

BankListSqlModel::BankListSqlModel(const QString &connectionName,
                                   ServerApi *api,
                                   IcoImageProvider *imageProvider,
                                   QSettings *settings)
    : ListSqlModel(connectionName, api, imageProvider, settings),
      mQuery(QSqlDatabase::database(connectionName)),
      mQueryUpdateBanks(QSqlDatabase::database(connectionName)),
      mQueryUpdateBankIco(QSqlDatabase::database(connectionName)),
      mQuerySetBankMine(QSqlDatabase::database(connectionName)),
      mQueryGetPartners(QSqlDatabase::database(connectionName)),
      mQuerySetPartners(QSqlDatabase::database(connectionName)),
      mQueryById(QSqlDatabase::database(connectionName)),
      mQueryGetMineBanks(QSqlDatabase::database(connectionName))
{
    setRowCount(600);

    setRoleName(IdRole,        "bank_id");
    setRoleName(NameRole,      "bank_name");
    setRoleName(LicenceRole,   "bank_licence");
    setRoleName(NameTrRole,    "bank_name_tr");
    setRoleName(RatingRole,    "bank_rating");
    setRoleName(NameTrAltRole, "bank_name_tr_alt");
    setRoleName(TelRole,       "bank_tel");
    setRoleName(MineRole,      "bank_is_mine");
    setRoleName(IcoPathRole,   "bank_ico_path");

    const QString queryErrPrefix = "BankListSqlModel cannot prepare query:";

    if (!mQuery.prepare("SELECT id, name, licence, name_tr, rating, town, name_tr_alt, tel, mine, ico_path FROM banks"
                        " WHERE"
                        "       name LIKE :name"
                        " or licence LIKE :licence"
                        " or name_tr LIKE :name_tr"
                        " or    town LIKE :town"
                        " or     tel LIKE :tel"
                        " ORDER BY mine DESC, rating ASC"))
    {
        qWarning() << queryErrPrefix << mQuery.lastError().databaseText();
    }

    if (!mQueryUpdateBanks.prepare("INSERT OR REPLACE INTO banks (id, name, licence, name_tr, rating, name_tr_alt, town, tel, mine)"
                                   "VALUES (:id, :name, :licence, :name_tr, :rating, :name_tr_alt, :town, :tel, :mine)"))
    {
        qWarning() << queryErrPrefix << mQueryUpdateBanks.lastError().databaseText();
    }

    if (!mQueryUpdateBankIco.prepare("UPDATE banks SET ico_path = :ico_path WHERE id = :bank_id")) {
        qWarning() << queryErrPrefix << mQueryUpdateBankIco.lastError().databaseText();
    }

    if (!mQuerySetBankMine.prepare("UPDATE banks SET mine = :mine WHERE id = :bank_id")) {
        qWarning() << queryErrPrefix << mQuerySetBankMine.lastError().databaseText();
    }

    if (!mQueryGetMineBanks.prepare("SELECT id FROM banks WHERE mine = 1")) {
        qWarning() << queryErrPrefix << mQueryGetMineBanks.lastError().databaseText();
    }

    if (!mQuerySetPartners.prepare("INSERT OR REPLACE INTO partners (id, partner_id)"
                                   "VALUES (:id, :partner_id)"))
    {
        qWarning() << queryErrPrefix << mQuerySetPartners.lastError().databaseText();
    }

    if (!mQueryGetPartners.prepare("SELECT partner_id FROM partners WHERE id = :id"))
    {
        qWarning() << queryErrPrefix << mQueryGetPartners.lastError().databaseText();
    }

    if (!mQueryById.prepare("SELECT id, name, licence, name_tr, rating, name_tr_alt, town, tel, mine "
                            "FROM banks WHERE id = :bank_id"))
    {
        qWarning() << queryErrPrefix << mQueryById.lastError().databaseText();
    }

    connect(this, SIGNAL(updateBanksIdsRequest(quint32)),
            this, SLOT(updateBanksIds(quint32)), Qt::QueuedConnection);

    connect(this, SIGNAL(bankIdsUpdated(quint32)),
            this, SLOT(restoreBanksData()), Qt::QueuedConnection);

    connect(this, SIGNAL(updateBanksDataRequest(quint32)),
            this, SLOT(updateBanksData(quint32)), Qt::QueuedConnection);

    connect(this, SIGNAL(updateBankIcoRequest(quint32,quint32)),
            this, SLOT(updateBankIco(quint32,quint32)), Qt::QueuedConnection);

    setFilter("", "{}");
}

QVariant BankListSqlModel::data(const QModelIndex &item, int role) const
{
    if (role < Qt::UserRole || role >= RoleLast)
    {
        return ListSqlModel::data(item, role);
    }

    return QStandardItemModel::data(index(item.row(), 0), role);
}

bool BankListSqlModel::setData(const QModelIndex &index,
                               const QVariant &value,
                               int role)
{
    if (role == MineRole) {
        const int mine = value.toInt() == 0 ? 0 : 1;
        quint32 bankId = index.data(IdRole).toInt();
        if (!bankId) {
            qDebug() << "invalid bank id";
            return false;
        }

        qDebug() << "update mine: [" << bankId << ": " << mine << "]";
        mQuerySetBankMine.bindValue(0, mine);
        mQuerySetBankMine.bindValue(1, bankId);
        if (!mQuerySetBankMine.exec()) {
            qDebug() << "BankListSqlModel cannot local update mine banks";
        }

        const QString bankIdStr = QString::number(bankId);

        getSettings()->beginGroup("mybanks");
        if (mine == 1) {
            getSettings()->setValue(bankIdStr, mine);
        } else {
            getSettings()->remove(bankIdStr);
        }
        getSettings()->endGroup();

        QJsonObject json;
        json["mine"] = QJsonValue(mine);
        /// TODO: add session
        getServerApi()->sendRequest("/bank/" + bankIdStr + "/mine", json,
        [](ServerApi::RequestStatusCode reqCode, ServerApi::HttpStatusCode code, const QByteArray &data) {

        });
    }
    return ListSqlModel::setData(index, value, role);
}

QList<int> BankListSqlModel::getMineBanks() const
{
    QList<int> result;
    if (!mQueryGetMineBanks.exec()) {
        qWarning() << "BankListSqlModel::getMineBanks query failed:" << mQueryGetMineBanks.lastError().databaseText();
        return result;
    }

    while (mQueryGetMineBanks.next()) {
        const int id = mQueryGetMineBanks.value(0).toInt();
        if (id > 0) {
            result.append(id);
        }
    }

    return result;
}

QList<int> BankListSqlModel::getPartnerBanks(int bankId)
{
    QList<int> result;
    mQueryGetPartners.bindValue(0, bankId);
    if (!mQueryGetPartners.exec()) {
        qWarning() << "BankListSqlModel::getPartnerBanks query failed:" << mQueryGetMineBanks.lastError().databaseText();
        return result;
    }

    while (mQueryGetPartners.next()) {
        const int id = mQueryGetPartners.value(0).toInt();
        if (id > 0) {
            result.append(id);
        }
    }

    return result;
}

QString BankListSqlModel::getBankData(int bankId) const
{
    mQueryById.bindValue(0, bankId);
    if (!mQueryById.exec()) {
        qWarning() << "BankListSqlModel::getBankData query failed:" << mQueryById.lastError().databaseText();
        return "";
    }

    if (mQueryById.next()) {
        int id = mQueryById.value(0).toInt();
        if (id > 0) {
            QJsonObject obj;

            obj["id"] = id;
            obj["name"] = mQueryById.value(1).toString();
            obj["licence"] = mQueryById.value(2).toInt();
            obj["name_tr"] = mQueryById.value(3).toString();
            obj["rating"] = mQueryById.value(4).toInt();
            obj["name_tr_alt"] = mQueryById.value(5).toString();
            /// TODO: obj["town"]
            obj["tel"] = mQueryById.value(7).toString();
            obj["mine"] = mQueryById.value(8).toBool();

            return QString::fromUtf8(QJsonDocument(obj).toJson(QJsonDocument::Compact));
        }
    }

    return "";
}

QList<int> BankListSqlModel::getPartnerBanks(const QList<int> &bankIdList)
{
    QSet<int> result;
    for (int id : bankIdList) {
        const QList<int> partnerList = getPartnerBanks(id);
        for (int partner : partnerList) {
            result.insert(partner);
        }
    }
    return result.toList();
}

void BankListSqlModel::updateFromServerImpl(quint32 leftAttempts)
{
    if (leftAttempts == 0) {
        return;
    }

    emitUpdateBanksIds(leftAttempts);
}

void BankListSqlModel::setFilterImpl(const QString &filter, const QJsonObject &options)
{
    Q_UNUSED(options);

    for (int i = 0; i < 5; ++i) {
        mQuery.bindValue(i, filter);
    }

    if (!mQuery.exec()) {
        qCritical() << mQuery.lastError().databaseText();
        Q_ASSERT_X(false, "TownListSqlModel::setFilter", "sql query error");
    }

    clear();
    int row = 0;
    while (mQuery.next())
    {
        QList<QStandardItem *> items;

        QStandardItem *item = new QStandardItem;
        for (int i = 0; i < RoleLast - IdRole; ++i) {
            item->setData(mQuery.value(i), IdRole + i);
//            qDebug() << mQuery.value(i);
        }
        items.append(item);

        insertRow(row, items);
        ++row;
    }
}

QList<int> BankListSqlModel::getSelectedIdsImpl() const
{
    QList<int> result;

    const int rows = rowCount();
    for (int i = 0; i < rows; i++) {
        QModelIndex idx = index(i, 0);
        if (data(idx, SelectedRole).toBool()) {
            result.append(data(idx, IdRole).toInt());
        }
    }

    return result;
}

static QList<int> getBanksIdList(const QJsonDocument &json)
{
    QList<int> bankIdList;

    if (!json.isArray()) {
        qWarning() << "getBankIdList: expected json int array";
        return bankIdList;
    }

    const QJsonArray arr = json.array();

    for (const QJsonValue &val : arr) {
        static const int invalidId = -1;
        const int id = val.toInt(invalidId);
        if (id > invalidId) {
            bankIdList.append(id);
        }
    }

    return bankIdList;
}

static QList<Bank> getBankList(const QJsonDocument &json)
{
    if (json.isArray()) {
        return Bank::fromJsonArray(json.array());
    }
    return {};
}

void BankListSqlModel::updateBanksIds(quint32 leftAttempts)
{
    if (leftAttempts == 0) {
        qDebug() << "updateBanksIds: no retry attempt left";
        emitRequestError(0, trUtf8("Could not connect to server after serval attempts"));
        return;
    }

    /// Get list of banks' ids
    const int requestId = getServerApi()->sendRequest("/banks", {},
    [&, requestId](ServerApi::RequestStatusCode reqCode, ServerApi::HttpStatusCode httpCode, const QByteArray &data) {
        if (reqCode == ServerApi::RSC_Timeout) {
            emitUpdateBanksIds(leftAttempts - 1);
            return;
        }

        if (reqCode != ServerApi::RSC_Ok) {
            emitRequestError(requestId, ServerApi::requestStatusCodeText(reqCode));
            return;
        }

        if (httpCode != ServerApi::HSC_Ok) {
            qWarning() << "Server request http code: " << httpCode;
            emitUpdateBanksIds(leftAttempts - 1);
            return;
        }

        QJsonParseError err;
        const QJsonDocument json = QJsonDocument::fromJson(data, &err);
        if (err.error != QJsonParseError::NoError) {
            emitRequestError(requestId, "updateBanksIds: server response json parse error: " + err.errorString());
            return;
        }

        mBanksToProcess = getBanksIdList(json);
        setUploadedCount(0);
        emitBankIdsUpdated(getAttemptsCount());
    });
}

void BankListSqlModel::updateBanksData(quint32 leftAttempts)
{
    if (leftAttempts == 0) {
        qDebug() << "updateBanksData: no retry attempt left";
        emitRequestError(0, trUtf8("Could not connect to server after serval attempts"));
        return;
    }

    const quint32 banksToProcess = qMin(getRequestBatchSize(), (quint32)mBanksToProcess.size());
    if (banksToProcess == 0) {
        saveInCache();
        emitServerDataReceived();
        return;
    }

    QJsonArray requestBanksBatch;
    for (quint32 i = 0; i < banksToProcess; ++i) {
        requestBanksBatch.append(mBanksToProcess.front());
        mBanksToProcess.removeFirst();
    }

    /// Get banks data from list
    const int requestId = getServerApi()->sendRequest("/banks", { QPair<QString, QJsonValue>("banks", requestBanksBatch) },
    [&, requestId](ServerApi::RequestStatusCode reqCode, ServerApi::HttpStatusCode httpCode, const QByteArray &data) {
        if (reqCode == ServerApi::RSC_Timeout) {
            for (const QJsonValue &val : requestBanksBatch) {
                const int id = val.toInt();
                if (id > 0) {
                    mBanksToProcess.append(id);
                }
            }

            emitUpdateBanksData(leftAttempts - 1);
            return;
        }

        if (reqCode != ServerApi::RSC_Ok) {
            emitRequestError(requestId, ServerApi::requestStatusCodeText(reqCode));
            return;
        }

        if (httpCode != ServerApi::HSC_Ok) {
            qWarning() << "updateBanksData: http status code: " << httpCode;
            for (const QJsonValue &val : requestBanksBatch) {
                const int id = val.toInt();
                if (id > 0) {
                    mBanksToProcess.append(id);
                }
            }

            emitUpdateBanksData(leftAttempts - 1);
            return;
        }

        QJsonParseError err;
        const QJsonDocument json = QJsonDocument::fromJson(data, &err);
        if (err.error != QJsonParseError::NoError) {
            qWarning() << "updateBanksData: response parse error: " << err.errorString();
            return;
        }

        const QList<Bank> bankList = getBankList(json);
        if (bankList.isEmpty()) {
            qWarning() << "updateBanksData: response is empty\n"
                       << QString::fromUtf8(data);
        }

        for (const Bank &bank : bankList) {
            mQueryUpdateBanks.bindValue(0, bank.id);
            mQueryUpdateBanks.bindValue(1, bank.name);
            mQueryUpdateBanks.bindValue(2, bank.licence);
            mQueryUpdateBanks.bindValue(3, bank.nameTr);
            mQueryUpdateBanks.bindValue(4, bank.rating);
            mQueryUpdateBanks.bindValue(5, bank.nameTrAlt);
            mQueryUpdateBanks.bindValue(6, bank.town);
            mQueryUpdateBanks.bindValue(7, bank.tel);

            if (bank.mine == 1) {
                getSettings()->beginGroup("mybanks");
                getSettings()->setValue(QString::number(bank.id), bank.mine);
                getSettings()->endGroup();
                mQueryUpdateBanks.bindValue(8, bank.mine);
            } else {
                getSettings()->beginGroup("mybanks");
                const int mine = getSettings()->value(QString::number(bank.id), 0).toInt();
                getSettings()->endGroup();
                mQueryUpdateBanks.bindValue(8, mine);
            }

            if (!mQueryUpdateBanks.exec()) {
                qWarning() << "updateBanksData: failed to update 'banks' table";
                qWarning() << "updateBanksData: " << mQueryUpdateBanks.lastError().databaseText();
            }

            if (!writeBankPartnersToDB(bank.id, bank.partners)) {
                qWarning() << "updateBanksData: failed to update 'partners' table";
                qWarning() << "updateBanksData: " << mQueryGetPartners.lastError().databaseText();
            }

            emitUpdateBankIco(getAttemptsCount(), bank.id);
        }

        setUploadedCount(getUploadedCount() + bankList.size());
        emitUpdateProgress();
        emitUpdateBanksData(getAttemptsCount());
    });
}

void BankListSqlModel::updateBankIco(quint32 leftAttempts, quint32 bankId)
{
    if (leftAttempts == 0) {
        qDebug() << "updateBankIco: no retry attempt left";
        emitRequestError(0, trUtf8("Could not connect to server after serval attempts"));
        return;
    }

    const QString requestPath = "/bank/" + QString::number(bankId) + "/ico";
    const int requestId = getServerApi()->sendRequest(requestPath, {},
    [&, bankId, requestId](ServerApi::RequestStatusCode reqCode, ServerApi::HttpStatusCode code, const QByteArray &data) {
        if (reqCode == ServerApi::RSC_Timeout) {
            emitUpdateBankIco(leftAttempts - 1, bankId);
            return;
        }

        if (reqCode != ServerApi::RSC_Ok) {
            emitRequestError(requestId, ServerApi::requestStatusCodeText(reqCode));
            return;
        }

        if (code != ServerApi::HSC_Ok) {
            qWarning() << "updateBankIco: http status code: " << code;

            emitUpdateBankIco(leftAttempts - 1, bankId);
            return;
        }

        QJsonParseError err;
        const QJsonDocument json = QJsonDocument::fromJson(data, &err);
        if (err.error != QJsonParseError::NoError) {
            emitRequestError(requestId, "updateBankIco: response parse error: " + err.errorString());
            return;
        }

        const QJsonObject jsonObj = json.object();
        if (jsonObj.isEmpty()) {
            emitRequestError(requestId, "updateBankIco: empty json object in response: " + data);
            return;
        }

        const int bankId = jsonObj["bank_id"].toInt();
        if (bankId == 0) {
            emitRequestError(requestId, "updateBankIco: invalid bankId");
            return;
        }

        QString icoData = jsonObj["ico_data"].toString();

        const QString bankIcoPath = "ico/bank/" + QString::number(bankId);
        if (!getIcoImageProvider()->loadSvgImage("bank/" + QString::number(bankId), icoData.toUtf8())) {
            emitRequestError(requestId, "updateBankIco: cannot load ico into ImageProvider: " + QString::number(bankId));
            return;
        }

        mQueryUpdateBankIco.bindValue(0, bankIcoPath);
        mQueryUpdateBankIco.bindValue(1, bankId);

        if (!mQueryUpdateBankIco.exec()) {
            QString msg = "failed to update 'banks' table. Ico for bank: " + QString::number(bankId) + " " +
                    mQueryUpdateBankIco.lastError().databaseText();
            emitRequestError(requestId, "updateBanksData: " + msg);
            return;
        }

        emitBankIcoUpdated(bankId);
    });
}

static QJsonDocument getBankListJson(const QList<Bank> &list)
{
    QJsonArray array;
    for (const Bank &bank : list) {
        array.append(QJsonValue(bank.toJsonObject()));
    }
    return QJsonDocument(array);
}

void BankListSqlModel::saveInCache()
{
    QSqlQuery query(QSqlDatabase::database(getDBConnectionName()));
    if (!query.exec("SELECT id, name, licence, name_tr, rating, name_tr_alt, town, tel, mine FROM banks"))
    {
        qWarning() << "Failed to save bank data cache due to sql error:" << query.lastError().databaseText();
        return;
    }

    QList<Bank> bankList;
    while (query.next()) {
        const int id = query.value(0).toInt();
        if (id > 0) {
            Bank bank;

            bank.id = id;
            bank.name = query.value(1).toString();
            bank.licence = query.value(2).toInt();
            bank.nameTr = query.value(3).toString();
            bank.rating = query.value(4).toInt();
            bank.nameTrAlt = query.value(5).toString();
            bank.town = query.value(6).toString();
            bank.tel = query.value(7).toString();

            bank.partners = getPartnerBanks(id);

            bankList.append(bank);
        }
    }

    QJsonDocument json = getBankListJson(bankList);

    QByteArray rawJson = json.toJson();
    QByteArray compressedJson = qCompress(rawJson);

    qDebug() << "Raw Bank data:" << rawJson.size();
    qDebug() << "Compressed data:" << compressedJson.size();

    const QString appDataPath = QStandardPaths::writableLocation(QStandardPaths::AppDataLocation);
    const QString bdataPath = QDir(appDataPath).absoluteFilePath("bdata");
    QFile dataFile(bdataPath);
    if (!dataFile.open(QIODevice::WriteOnly)) {
        qWarning("Cannot open tdata file for writing!");
        return;
    }

    dataFile.write(compressedJson);
}

void BankListSqlModel::restoreBanksData()
{
    //restoreFromCache(mBanksToProcess);
    setExpectedUploadCount(mBanksToProcess.size());
    if (!mBanksToProcess.isEmpty()) { // fetch left banks from server
        emitUpdateBanksData(getAttemptsCount());
    } else { // all towns restored from cache
        emitServerDataReceived();
    }
}

void BankListSqlModel::restoreFromCache(QList<int> &bankIdList)
{
    const QString bdataPath = QStandardPaths::locate(QStandardPaths::AppDataLocation, "bdata");
    if (bdataPath.isEmpty()) {
        return;
    }

    QFile dataFile(bdataPath);
    if (!dataFile.open(QIODevice::ReadOnly)) {
        qWarning("Cannot open bdata file for reading!");
        return;
    }

    QFileInfo finfo(dataFile);
    const QDateTime lastModified = finfo.lastModified();
    const qint64 daySecs = 3600 * 24;

    qDebug() << "last bdata modified time:" << lastModified;

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
        qWarning() << "Cannot decode bdata json!" << err.errorString();

        QFile rmFile(bdataPath);
        rmFile.open(QIODevice::Truncate);
        return;
    }

    //qDebug() << "Raw data has been read:" << rawJson;

    if (!json.isArray()) {
        qWarning() << "Json array expected in bdata file!";

        QFile rmFile(bdataPath);
        rmFile.open(QIODevice::Truncate);
        return;
    }

    QList<Bank> bankList = getBankList(json);
    qDebug() << "Restored banks from cache:" << bankList.size();

    for (auto it = bankList.begin(); it != bankList.end(); it++) {
        bool removed = bankIdList.removeOne((int)it->id);
        if (!removed) { // no such bank in server db
            it = bankList.erase(it);
        } else { // bank exists in server db
            emitUpdateBankIco(getAttemptsCount(), it->id);
            writeBankToDB(*it);
        }
    }

//    QSqlQuery query(QSqlDatabase::database(getDBConnectionName()));
//    if (!query.exec("SELECT COUNT(*) FROM banks")) {
//        qDebug() << query.lastError().databaseText();
//    }
//    if (query.next()) {
//        qDebug() << "Banks count: " << query.value(0).toInt();
//    } else {
//        qDebug() << "Cannot fetch banks count!";
//    }
}

void BankListSqlModel::writeBankToDB(const Bank &bank)
{
    mQueryUpdateBanks.bindValue(0, bank.id);
    mQueryUpdateBanks.bindValue(1, bank.name);
    mQueryUpdateBanks.bindValue(2, bank.licence);
    mQueryUpdateBanks.bindValue(3, bank.nameTr);
    mQueryUpdateBanks.bindValue(4, bank.rating);
    mQueryUpdateBanks.bindValue(5, bank.nameTrAlt);
    mQueryUpdateBanks.bindValue(6, bank.town);
    mQueryUpdateBanks.bindValue(7, bank.tel);
    mQueryUpdateBanks.bindValue(8, bank.mine);

    if (!mQueryUpdateBanks.exec()) {
        qWarning() << "writeBankToDB: failed to update 'banks' table";
        qWarning() << "writeBankToDB: " << mQueryUpdateBanks.lastError().databaseText();
        return;
    }

    if (!writeBankPartnersToDB(bank.id, bank.partners)) {
        qWarning() << "writeBankToDB: failed to update 'partners' table";
        qWarning() << "writeBankToDB: " << mQuerySetPartners.lastError().databaseText();
    }
}

bool BankListSqlModel::writeBankPartnersToDB(quint32 id, const QList<int> &partners)
{
    bool ok = true;
    for (int partnerId : partners) {
        mQuerySetPartners.bindValue(0, id);
        mQuerySetPartners.bindValue(1, partnerId);
        if (!mQuerySetPartners.exec()) {
            ok = false;
        }
    }
    return ok;
}
