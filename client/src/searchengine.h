#ifndef SEARCHENGINE_H
#define SEARCHENGINE_H

#include "listsqlmodel.h"

class BankListSqlModel;
class TownListSqlModel;

class SearchEngineFilter;

class SearchEngine : public ListSqlModel
{
    Q_OBJECT

    Q_PROPERTY(int rowCount READ rowCount NOTIFY rowCountChanged)
    Q_PROPERTY(int suggestionsCount READ getSuggestionsCount WRITE setSuggestionsCount NOTIFY suggestionsCountChanged)
    Q_PROPERTY(bool showOnlyMineBanks READ isShowingOnlyMineBanks WRITE showOnlyMineBanks NOTIFY showOnlyMineBanksChanged)
    Q_PROPERTY(bool showOnlyApprovedPoints READ isShowingOnlyApprovedPoints WRITE showOnlyApprovedPoints NOTIFY showOnlyApprovedPointsChanged)
    //Q_PROPERTY(QString filterPatch READ getFilterPatch WRITE setFilterPatch NOTIFY filterPatchChanged)
    Q_PROPERTY(bool showPartnerBanks READ isShowingPartnerBanks WRITE setShowPartnerBanks NOTIFY showPartnerBanksChanged)

public:
    enum Roles {
        IdRole = Qt::UserRole,
        NameRole,
        TypeRole,
        IcoRole,
        LongituteRole,
        LatitudeRole,
        ZoomRole,
        CandidateRole,
        FilterPatchRole,

        RoleLast
    };

    SearchEngine(BankListSqlModel *bankListModel,
                 TownListSqlModel *townListModel);

    ~SearchEngine();

    Q_INVOKABLE QString getCandidate() const;
    inline int getSuggestionsCount() const { return mSuggestionsCount; }

    Q_INVOKABLE QString getMineBanksFilter();

    inline bool isShowingOnlyMineBanks() const { return mShowOnlyMineBanks; }
    inline bool isShowingOnlyApprovedPoints() const { return mShowOnlyApprovedPoints; }
    inline bool isShowingPartnerBanks() const { return mShowPartnerBanks; }

signals:
    void rowCountChanged(int count);
    void suggestionsCountChanged(int count);    
    void showOnlyMineBanksChanged(bool enabled);
    void showOnlyApprovedPointsChanged(bool enabled);
    void showPartnerBanksChanged(bool showing);

public slots:
    void setSuggestionsCount(int count);
    void showOnlyMineBanks(bool enabled);
    void showOnlyApprovedPoints(bool enabled);

    void setShowPartnerBanks(bool show);

protected:
    bool setData(const QModelIndex &index, const QVariant &value, int role);

    void setFilterImpl(const QString &filter, const QJsonObject &options) override;
    int getLastRole() const override { return RoleLast; }
    bool needEscapeFilter() const override { return false; }
    void updateFromServerImpl(quint32) override { }

    QList<int> getSelectedIdsImpl() const override { return {}; }

    QSqlQuery *getQuery() override { return nullptr; }

private:
    static void fillJsonData(const QStandardItem *item, QJsonObject &json);

    struct FilterPatch {
        QString name;
//        QJsonObject filter;
    };

    QList<FilterPatch> mSavedRequests;

    QList<SearchEngineFilter *> mFilters;

    BankListSqlModel *mBankListModel;
    TownListSqlModel *mTownListModel;

    int mSuggestionsCount;
    bool mShowOnlyMineBanks;
    bool mShowOnlyApprovedPoints;
    bool mShowPartnerBanks;
};

#endif // SEARCHENGINE_H
