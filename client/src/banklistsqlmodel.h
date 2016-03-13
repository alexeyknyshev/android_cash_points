#ifndef BANKLISTSQLMODEL_H
#define BANKLISTSQLMODEL_H

#include <QtSql/QSqlQuery>

#include "listsqlmodel.h"

class Bank;

class BankListSqlModel : public ListSqlModel
{
    Q_OBJECT

    friend class SearchEngine;

public:
    enum Roles {
        IdRole = Qt::UserRole,
        NameRole,
        LicenceRole,
        NameTrRole,
        RatingRole,
        RegionName,
        NameTrAltRole,
        TelRole,
        MineRole,
        IcoPathRole,

        RoleLast
    };

    BankListSqlModel(const QString &connectionName,
                     ServerApi *api,
                     IcoImageProvider *imageProvider,
                     QSettings *settings);

    QVariant data(const QModelIndex &item, int role) const override;

    bool setData(const QModelIndex &index,
                 const QVariant &value,
                 int role) override;

    Q_INVOKABLE QList<int> getMineBanks() const;

    Q_INVOKABLE QList<int> getPartnerBanks(const QList<int> &bankIdList);
    QList<int> getPartnerBanks(int bankId);

    Q_INVOKABLE QString getBankData(int bankId) const;

    QSqlQuery *getQuery() override { return &mQuery; }

signals:
    void updateBanksIdsRequest(quint32 leftAttempts);
    void updateBanksDataRequest(quint32 leftAttempts);
    void updateBankIcoRequest(quint32 leftAttempts, quint32 bankId);

    void bankIdsUpdated(quint32 leftAttempts);
    void bankIcoUpdated(quint32 bankId);

protected:
    void updateFromServerImpl(quint32 leftAttempts) override;
    void setFilterImpl(const QString &filter) override;

    int getLastRole() const override { return RoleLast; }
    bool needEscapeFilter() const override { return true; }


private slots:
    void restoreBanksData();
    void updateBanksIds(quint32 leftAttempts);
    void updateBanksData(quint32 leftAttempts);
    void updateBankIco(quint32 leftAttempts, quint32 bankId);

private:
    void emitUpdateBanksIds(quint32 leftAttempts)
    { emit updateBanksIdsRequest(leftAttempts); }
    void emitBankIdsUpdated(quint32 leftAttempts)
    { emit bankIdsUpdated(leftAttempts); }
    void emitUpdateBanksData(quint32 leftAttempts)
    { emit updateBanksDataRequest(leftAttempts); }
    void emitUpdateBankIco(quint32 leftAttempts, quint32 bankId)
    { emit updateBankIcoRequest(leftAttempts, bankId); }
    void emitBankIcoUpdated(quint32 bankId)
    { emit bankIcoUpdated(bankId); }

    void saveInCache();
    void restoreFromCache(QList<int> &bankIdList);

    void writeBankToDB(const Bank &bank);
    bool writeBankPartnersToDB(quint32 id, const QList<int> &partners);

    QList<int> mBanksToProcess;

    QSqlQuery mQuery;
    QSqlQuery mQueryUpdateBanks;
    QSqlQuery mQueryUpdateBankIco;
    QSqlQuery mQuerySetBankMine;
    QSqlQuery mQueryGetPartners;
    QSqlQuery mQuerySetPartners;
    mutable QSqlQuery mQueryById;
    mutable QSqlQuery mQueryGetMineBanks;
};

#endif // BANKLISTSQLMODEL_H
