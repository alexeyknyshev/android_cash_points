#include "searchengine.h"

#include "banklistsqlmodel.h"
#include "townlistsqlmodel.h"

#include <QtSql/QSqlQuery>
#include <QtSql/QSqlError>
#include <QtCore/QJsonDocument>
#include <QtCore/QJsonObject>
#include <QtCore/QJsonArray>
#include <QtCore/QDebug>
#include <QtCore/QSettings>

class SearchEngineFilter
{
public:
    virtual ~SearchEngineFilter() { }

    class Suggestion
    {
    public:
        Suggestion(bool isDecorator,
                   QList<QString> var = {},
                   const QJsonObject &patch = QJsonObject(),
                   const QMap<int, QVariant> &rolesData = {})
            : mIsDecorator(isDecorator),
              mVariants(var),
              mPatch(patch),
              mRolesData(rolesData)
        { }

        Suggestion(const Suggestion &other)
            : mIsDecorator(other.isDecorator()),
              mVariants(other.getVariants()),
              mPatch(other.getJsonPatch()),
              mRolesData(other.getRolesData())
        { }

        void operator=(const Suggestion &other)
        {
            mIsDecorator = other.isDecorator();
            mVariants = other.getVariants();
            mPatch = other.getJsonPatch();
            mRolesData = other.getRolesData();
        }

        bool isDecorator() const { return mIsDecorator; }

        const QStringList &getVariants() const { return mVariants; }

        const QMap<int, QVariant> &getRolesData() const { return mRolesData; }

        const QJsonObject &getJsonPatch() const { return mPatch; }

        void join(const Suggestion &other) {
            if (isDecorator()) {
                Q_ASSERT_X(false, "Suggestion::join(const Suggestion&)", "Attempt to join on decorator");
                qWarning() << "Suggestion::join(const Suggestion&): Attempt to join on decorator";
            }

            if (!other.isDecorator()) {
                Q_ASSERT_X(false, "Suggestion::join(const Suggestion&)", "Attempt to join with non decorator");
                qWarning() << "Suggestion::join(const Suggestion&): Attempt to join with non decorator";
            }

            mVariants.append(other.getVariants());

            const auto end = other.getJsonPatch().constEnd();
            for (auto it = other.getJsonPatch().constBegin(); it != end; it++) {
                mPatch.insert(it.key(), it.value());
            }
        }

    private:
        bool mIsDecorator;
        QStringList mVariants;
        QJsonObject mPatch;
        QMap<int, QVariant> mRolesData;
    };

    virtual QList<Suggestion> filter(QString &request) const = 0;

protected:
    static bool removeMatching(QString &request, const QString &matching)
    {
        const QString requestLower = request.toLower();
        int index = requestLower.indexOf(matching);
        if (index != -1) {
            int matchEnd = requestLower.indexOf(' ', index + matching.size());
            if (matchEnd == -1) {
                matchEnd = requestLower.size();
            }
            request.remove(index, matchEnd - index);
            return true;
        }
        return false;
    }
};

class Filter24Hour : public SearchEngineFilter
{
public:
    virtual QList<SearchEngineFilter::Suggestion> filter(QString &request) const override {
        QList<QString> matching = {
            "24",
            "круглосуточ",
            "без перерыва",
            "round",
            "around the clock",
            "day and night",
        };

        for (const QString &m : matching) {
            if (removeMatching(request, m)) {
                return { Suggestion(true,
                                   { QObject::trUtf8("круглосуточно") },
                                   QJsonObject{ { "round_the_clock", true } }) };
            }
        }

        return { };
    }
};

class FilterBanks : public SearchEngineFilter
{
public:
    FilterBanks(BankListSqlModel *model)
        : mModel(model)
    { }

    virtual QList<SearchEngineFilter::Suggestion> filter(QString &request) const override {
        request = request.trimmed();
        const QString escapedRequest = mModel->escapeFilter(request);

        QSqlQuery *bankQuery = mModel->getQuery();
        Q_ASSERT_X(bankQuery, "FilterBanks::filter(QString &)", "null bankQuery ptr");

        bankQuery->bindValue(":name", escapedRequest);
        bankQuery->bindValue(":licence", "");
        bankQuery->bindValue(":name_tr", escapedRequest);
        bankQuery->bindValue(":town", "");
        bankQuery->bindValue(":tel", "");
        if (!bankQuery->exec()) {
            qWarning() << "SearchEngine::setFilter:" << bankQuery->lastError().databaseText();
            return {};
        }

        QList<Suggestion> suggestions;
        while (bankQuery->next()) {
            const int bankId = bankQuery->value(0).toInt();
            const QString bankName = bankQuery->value(1).toString();

            Suggestion sug(false,
                           { bankName },
                           QJsonObject{
                               { "bank_id", QJsonArray({ bankId }) }
                           },
                           QMap<int, QVariant>{
                               { SearchEngine::IdRole, bankId },
                               { SearchEngine::IcoRole, bankQuery->value(9) },
                               { SearchEngine::TypeRole, "bank" },
                           });
            suggestions.append(sug);
        }
        return suggestions;
    }
private:
    BankListSqlModel *mModel;
};

class FilterTowns : public SearchEngineFilter
{
public:
    FilterTowns(TownListSqlModel *model)
        : mModel(model)
    { }

    virtual QList<SearchEngineFilter::Suggestion> filter(QString &request) const override {
        request = request.trimmed();
        const QString escapedRequest = mModel->escapeFilter(request);

        QSqlQuery *townQuery = mModel->getQuery();
        Q_ASSERT_X(townQuery, "FilterTowns::filter(QString &)", "null townQuery ptr");

        townQuery->bindValue(":name", escapedRequest);
        townQuery->bindValue(":name_tr", escapedRequest);
        townQuery->bindValue(":region_name", "");
        if (!townQuery->exec()) {
            qWarning() << "SearchEngine::setFilter:" << townQuery->lastError().databaseText();
            return {};
        }

        QList<Suggestion> suggestions;
        while (townQuery->next()) {
            const int townId = townQuery->value(0).toInt();
            const QString townName = townQuery->value(1).toString();

            Suggestion sug(false,
                           { townName },
                           QJsonObject(),
                           QMap<int, QVariant>{
                               { SearchEngine::IdRole, townId },
                               { SearchEngine::LongituteRole, townQuery->value(4) },
                               { SearchEngine::LatitudeRole, townQuery->value(5) },
                               { SearchEngine::ZoomRole, townQuery->value(6) },
                               { SearchEngine::TypeRole, "town" },
                               { SearchEngine::IcoRole, "" },
                           });
            suggestions.append(sug);
        }
        return suggestions;
    }

private:
    TownListSqlModel *mModel;
};

class FilterPointType : public SearchEngineFilter
{
public:
    virtual QList<SearchEngineFilter::Suggestion> filter(QString &request) const override {
        {
            QList<QString> matchingAtm = {
                "банкомат",
                "терминал",
                "atm",
            };

            for (const QString &m : matchingAtm) {
                if (removeMatching(request, m)) {
                    return { Suggestion(true,
                                        { QObject::trUtf8("банкомат") },
                                        QJsonObject{ { "type", "atm" } }) };
                }
            }
        }

        {
            QList<QString> matchingOffice = {
                "офис",
                "отделение",
                "office",
                "branch",
            };

            for (const QString &m : matchingOffice) {
                if (removeMatching(request, m)) {
                    return { Suggestion(true,
                                        { QObject::trUtf8("офис") },
                                        QJsonObject{ { "type", "office" } }) };
                }
            }
        }

        return { };
    }
};

class FilterCurrency : public SearchEngineFilter
{
public:
    virtual QList<SearchEngineFilter::Suggestion> filter(QString &request) const override {
        {
            QList<QString> matchingRub = {
                "рубли",
                "рубль",
                "rub",
                "rubles",
            };

            for (const QString &m : matchingRub) {
                if (removeMatching(request, m)) {
                    return { Suggestion(true,
                                        { QObject::trUtf8("рубли") },
                                        QJsonObject{ { "rub", true } },
                                        QMap<int, QVariant>{
                                            { SearchEngine::NameRole, QObject::trUtf8("Валюта: рубли") },
                                            { SearchEngine::TypeRole, "currency" },
                                        }) };
                }
            }
        }

        {
            QList<QString> matchingUsd = {
                "доллар",
                "usd",
            };

            for (const QString &m : matchingUsd) {
                if (removeMatching(request, m)) {
                    return { Suggestion(true,
                                        { QObject::trUtf8("доллары") },
                                        QJsonObject{ { "usd", true } },
                                        QMap<int, QVariant>{
                                            { SearchEngine::NameRole, QObject::trUtf8("Валюта: доллары США") },
                                            { SearchEngine::TypeRole, "currency" },
                                        }) };
                }
            }
        }

        {
            QList<QString> matchingEur = {
                "евро",
                "eur",
            };

            for (const QString &m : matchingEur) {
                if (removeMatching(request, m)) {
                    return { Suggestion(true,
                                        { QObject::trUtf8("евро") },
                                        QJsonObject{ { "eur", true } },
                                        QMap<int, QVariant>{
                                            { SearchEngine::NameRole, QObject::trUtf8("Валюта: евро") },
                                            { SearchEngine::TypeRole, "currency" },
                                        }) };
                }
            }
        }

        return {};
    }
};

#define MY_BANKS "mybanks"
#define SHOW_PARTNER_BANKS "showPartnerBanks"

SearchEngine::SearchEngine(BankListSqlModel *bankListModel,
                           TownListSqlModel *townListModel)
    : ListSqlModel(bankListModel),
      mBankListModel(bankListModel),
      mTownListModel(townListModel)
{
    mFilters.append(new Filter24Hour);
    mFilters.append(new FilterPointType);
    mFilters.append(new FilterCurrency);
    mFilters.append(new FilterBanks(bankListModel));
    mFilters.append(new FilterTowns(townListModel));

    setSuggestionsCount(5);
    showOnlyMineBanks(false);

    getSettings()->beginGroup(MY_BANKS);
    mShowPartnerBanks = getSettings()->value(SHOW_PARTNER_BANKS, false).toBool();
    getSettings()->endGroup();

    setRoleName(IdRole,   "id");
    setRoleName(NameRole, "text");
    setRoleName(TypeRole, "type");
    setRoleName(IcoRole,  "ico");
    setRoleName(LongituteRole, "longitude");
    setRoleName(LatitudeRole, "latitude");
    setRoleName(ZoomRole, "zoom");
    setRoleName(CandidateRole, "candidate");
    setRoleName(FilterPatchRole, "filter");
}

SearchEngine::~SearchEngine()
{
    for (SearchEngineFilter *filter : mFilters) {
        delete filter;
    }
    mFilters.clear();
}

bool SearchEngine::setData(const QModelIndex &index, const QVariant &value, int role)
{
    if (role != CandidateRole) {
        return false;
    }

    QStandardItem *it = itemFromIndex(index);
    if (!it) {
        return false;
    }

    for (int row = 0; row < rowCount(); row++) {
        QStandardItem *currentItem = item(row);
        if (currentItem && currentItem->data(CandidateRole).toBool()) {
            currentItem->setData(false, CandidateRole);
        }
    }

    it->setData(value, role);
    qDebug() << "Candidate:" << it->data(NameRole).toString();
    return true;
}

void SearchEngine::fillJsonData(const QStandardItem *item, QJsonObject &json)
{
    json["id"] = item->data(IdRole).toInt();
    json["name"] = item->data(NameRole).toString();

    const QString type = item->data(TypeRole).toString();
    json["type"] = type;
    if (type == "town") {
        json["longitude"] = item->data(LongituteRole).toReal();
        json["latitude"] = item->data(LatitudeRole).toReal();
        json["zoom"] = item->data(ZoomRole).toReal();
    }

    const QString filterPatch = item->data(FilterPatchRole).toString();
    if (!filterPatch.isEmpty()) {
        QJsonParseError err;
        QJsonObject patchObj = QJsonDocument::fromJson(filterPatch.toUtf8(), &err).object();
        if (err.error != QJsonParseError::NoError) {
            qWarning() << "SearchEngine::fillJsonData(const QStandardItem *, QJsonObject &):"
                       << "cannot decode filter patch json";
        } else {
            json["filter"] = patchObj;
        }
    }
}

QString SearchEngine::getCandidate() const
{
    for (int row = 0; row < rowCount(); row++) {
        QStandardItem *currentItem = item(row);
        if (currentItem->data(CandidateRole).toBool() == true) {
            const int id = currentItem->data(IdRole).toInt();
            if (id > 0) {
                QJsonObject json;
                fillJsonData(currentItem, json);
                return QString::fromUtf8(QJsonDocument(json).toJson());
            }
        }
    }

    return "";
}

QString SearchEngine::getMineBanksFilter()
{
    QJsonObject obj;

    if (isShowingOnlyMineBanks()) {
        QList<int> mineBanks = mBankListModel->getMineBanks();
        if (!mineBanks.isEmpty()) {
            QJsonArray bankIdList;
            for (int id : mineBanks) {
                bankIdList.append(QJsonValue(id));
            }
            obj["bank_id"] = bankIdList;
        }
    }

//    QJsonParseError err;
//    QJsonObject filterPatch = QJsonDocument::fromJson(getFilterPatch().toUtf8(), &err).object();
//    if (err.error != QJsonParseError::NoError) {
//        qWarning() << "SearchEngine::getMineBanksFilter():"
//                   << "cannot decode filter patch json";
//        setFilterPatch("");
//    } else {
//        const auto end = filterPatch.constEnd();
//        for (auto it = filterPatch.constBegin(); it != end; ++it) {
//            obj.insert(it.key(), it.value());
//        }
//    }

    QJsonDocument json;
    json.setObject(obj);

    return QString::fromUtf8(json.toJson(QJsonDocument::Compact));
}

void SearchEngine::setFilterImpl(const QString &filter)
{
    clear();
    if (filter.isEmpty() || filter.count('%') == filter.size()) {
        emit rowCountChanged(0);
        return;
    }

    qDebug() << "SearchEngine::setFilterImpl:" << filter;

    QList<QStandardItem *> itemBatch;
    QString filterCopy = filter;

    QList<SearchEngineFilter::Suggestion> suggestions;
    QList<SearchEngineFilter::Suggestion> decorators;

    for (const SearchEngineFilter *f : mFilters) {
        QList<SearchEngineFilter::Suggestion> sugList = f->filter(filterCopy);
        if (!sugList.isEmpty()) {
            if (sugList.first().isDecorator()) {
                decorators.append(sugList);
            } else {
                suggestions.append(sugList);
            }
        }
/*
        if (sug.isValid() && sug.isDecorator()) {
            QStandardItem *item = new QStandardItem;
            QJsonObject json;
            item->setData(sug, NameRole);
            item->setData(QString::fromUtf8(QJsonDocument(json).toJson()), FilterPatchRole);
            item->setData("", IcoRole);
            item->setData("other", TypeRole);
            if (itemBatch.isEmpty()) {
                item->setData(true, CandidateRole);
            }
            itemBatch.append(item);

            const auto end = json.constEnd();
            for (auto it = json.constBegin(); it != end; ++it) {
                jsonPatch.insert(it.key(), it.value());
            }
        }
*/
    }

    {
        const auto end = suggestions.end();
        for (auto it = suggestions.begin(); it != end; it++) {
            for (const SearchEngineFilter::Suggestion &dec : decorators) {
                it->join(dec);
            }
        }
    }
    {
        const auto end = suggestions.cend();
        int suggestionIndex = 0;
        for (auto it = suggestions.cbegin();
             it != end && suggestionIndex < getSuggestionsCount();
             it++, suggestionIndex++)
        {
            QStandardItem *item = new QStandardItem;
            const SearchEngineFilter::Suggestion &sug = *it;

            item->setData(sug.getVariants().join(' '), SearchEngine::NameRole);
            const auto rolesEnd = sug.getRolesData().cend();
            for (auto rolesIt = sug.getRolesData().cbegin(); rolesIt != rolesEnd; rolesIt++) {
                item->setData(rolesIt.value(), rolesIt.key());
            }
            if (!sug.getJsonPatch().isEmpty()) {
                QString json = QString::fromUtf8(QJsonDocument(sug.getJsonPatch()).toJson());
                item->setData(json, SearchEngine::FilterPatchRole);
            }
            itemBatch.append(item);
        }
    }
/*
    filterCopy = filterCopy.trimmed();
    qDebug() << "filter:" << filterCopy;
    if (!filterCopy.isEmpty()) {
        filterCopy = escapeFilter(filterCopy);

        QSqlQuery *townQuery = mTownListModel->getQuery();
        Q_ASSERT_X(townQuery, "SearchEngine::setFilterImpl(const QString &)", "null townQuery ptr");

        townQuery->bindValue(":name", filterCopy);
        townQuery->bindValue(":name_tr", filterCopy);
        townQuery->bindValue(":region_name", "");
        if (!townQuery->exec()) {
            qWarning() << "SearchEngine::setFilter:" << townQuery->lastError().databaseText();
            return;
        }

        QSqlQuery *bankQuery = mBankListModel->getQuery();
        Q_ASSERT_X(bankQuery, "SearchEngine::setFilterImpl(const QString &)", "null bankQuery ptr");

        bankQuery->bindValue(":name", filterCopy);
        bankQuery->bindValue(":licence", "");
        bankQuery->bindValue(":name_tr", filterCopy);
        bankQuery->bindValue(":town", "");
        bankQuery->bindValue(":tel", "");
        if (!bankQuery->exec()) {
            qWarning() << "SearchEngine::setFilter:" << bankQuery->lastError().databaseText();
            return;
        }

        while (townQuery->next() && itemBatch.size() < mSuggestionsCount) {
            QStandardItem *item = new QStandardItem;
            item->setData(townQuery->value(0), IdRole);
            item->setData(townQuery->value(1), NameRole);
            item->setData(townQuery->value(4), LongituteRole);
            item->setData(townQuery->value(5), LatitudeRole);
            item->setData(townQuery->value(6), ZoomRole);
            item->setData("", IcoRole);
            item->setData("town", TypeRole);
            if (itemBatch.isEmpty()) {
                item->setData(true, CandidateRole);
            }
            qDebug() << townQuery->value(1);
            itemBatch.append(item);
        }
        while (bankQuery->next()) {
            QStandardItem *item = new QStandardItem;
            const int bankId = bankQuery->value(0).toInt();
            QString bankName = bankQuery->value(1).toString();
            if (!bankName.isEmpty() && !suggestion.isEmpty()) {
                bankName += " " + suggestion;
            }

            item->setData(bankId, IdRole);
            item->setData(bankName, NameRole);
            item->setData(bankQuery->value(9), IcoRole);
            item->setData("bank", TypeRole);

            jsonPatch["bank_id"] = QJsonArray({ bankId });
            item->setData(QString::fromUtf8(QJsonDocument(jsonPatch).toJson()), FilterPatchRole);

            if (itemBatch.isEmpty()) {
                item->setData(true, CandidateRole);
            }
            qDebug() << bankQuery->value(1);

            if (itemBatch.size() == mSuggestionsCount) {
                delete itemBatch.takeLast();
                itemBatch.append(item);
                break;
            }
            itemBatch.append(item);
        }
    }*/

    if (!itemBatch.isEmpty()) {
        itemBatch.first()->setData(true, SearchEngine::CandidateRole);
    }

    for (QStandardItem *item : itemBatch) {
        appendRow(item);
    }

    qDebug() << "Row count changed:" << itemBatch.size();
    emit rowCountChanged(itemBatch.size());
}

void SearchEngine::setSuggestionsCount(int count)
{
    if (count <= 0) {
        qFatal("SearchEngine: atempt to set non positive row count");
        return;
    }
    mSuggestionsCount = count;
    emit suggestionsCountChanged(count);
}

void SearchEngine::showOnlyMineBanks(bool enabled)
{
    mShowOnlyMineBanks = enabled;
    emit showOnlyMineBanksChanged(enabled);
}

void SearchEngine::setFilterPatch(QString filterPatch)
{
    mFilterPatch = filterPatch;
    emit filterPatchChanged(filterPatch);
}

void SearchEngine::setShowPartnerBanks(bool show)
{
    if (show != mShowPartnerBanks) {
        mShowPartnerBanks = show;
        getSettings()->beginGroup(MY_BANKS);
        getSettings()->setValue(SHOW_PARTNER_BANKS, show);
        getSettings()->endGroup();
        emit showPartnerBanksChanged(show);
    }
}
