#ifndef BANKLISTSQLMODEL_H
#define BANKLISTSQLMODEL_H

#include <QtSql/QSqlQuery>

#include "listsqlmodel.h"

class ServerApi;

class BankListSqlModel : public ListSqlModel
{
    Q_OBJECT

public:
    enum Roles {
        IdRole = Qt::UserRole,
        NameRole,
        LicenceRole,
        NameTrRole,
        RaitingRole,
        NameTrAltRole,
        TelRole,

        RoleLast
    };

    explicit BankListSqlModel(QString connectionName, ServerApi *api);

    QVariant data(const QModelIndex &item, int role) const override;

signals:
    void updateBanksDataRequest(quint32 leftAttempts);
    void syncNextBankBatch(quint32 leftAttempts);

protected:
    QHash<int, QByteArray> roleNames() const override;
    void updateFromServerImpl(quint32 leftAttempts) override;
    void setFilterImpl(const QString &filter) override;

    int getLastRole() const override { return RoleLast; }

private slots:
    void syncBanks(quint32 leftAttempts);

private:
    void emitUpdateBanksData(quint32 leftAttempts)
    { emit updateBanksDataRequest(leftAttempts); }

    QList<int> mBanksToProcess;

    QHash<int, QByteArray> mRoleNames;
    QString mQueryMask;
    QSqlQuery mQuery;
};

#endif // BANKLISTSQLMODEL_H
