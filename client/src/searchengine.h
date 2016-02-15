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
    Q_PROPERTY(QString filterPatch READ getFilterPatch WRITE setFilterPatch NOTIFY filterPatchChanged)
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
    inline const QString &getFilterPatch() const { return mFilterPatch; }

    Q_INVOKABLE QString getMineBanksFilter();

    inline bool isShowingOnlyMineBanks() const { return mShowOnlyMineBanks; }
    inline bool isShowingPartnerBanks() const { return mShowPartnerBanks; }

signals:
    void rowCountChanged(int count);
    void suggestionsCountChanged(int count);    
    void showOnlyMineBanksChanged(bool enabled);
    void filterPatchChanged(QString filterPatch);
    void showPartnerBanksChanged(bool showing);

public slots:
    void setSuggestionsCount(int count);
    void showOnlyMineBanks(bool enabled);
    void setFilterPatch(QString filterPatch);
    void setShowPartnerBanks(bool show);

protected:
    bool setData(const QModelIndex &index, const QVariant &value, int role);

    void setFilterImpl(const QString &filter);
    int getLastRole() const override { return RoleLast; }
    bool needEscapeFilter() const override { return false; }
    void updateFromServerImpl(quint32) override { }

    QSqlQuery *getQuery() override { return nullptr; }

private:
    static void fillJsonData(const QStandardItem *item, QJsonObject &json);

    QList<SearchEngineFilter *> mFilters;

    BankListSqlModel *mBankListModel;
    TownListSqlModel *mTownListModel;

    int mSuggestionsCount;
    bool mShowOnlyMineBanks;
    bool mShowPartnerBanks;
    QString mFilterPatch;
};

#endif // SEARCHENGINE_H
