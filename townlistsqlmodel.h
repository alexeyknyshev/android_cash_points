#ifndef TOWNLISTSQLMODEL_H
#define TOWNLISTSQLMODEL_H

#include <QtGui/QStandardItemModel>
#include <QtSql/QSqlQuery>

class TownListSqlModel : public QStandardItemModel
{
    Q_OBJECT

public:
    enum Roles {
        IdRole = Qt::UserRole,
        NameRole,
        NameTrRole,
        RegionRole
    };

    explicit TownListSqlModel(QString connctionName);

    QVariant data(const QModelIndex &item, int role) const override;

public slots:
    void setFilter(QString filterStr);

protected:
    QHash<int, QByteArray> roleNames() const override;

private:
    QHash<int, QByteArray> mRoleNames;
    QSqlQuery mQuery;
};

#endif // TOWNLISTSQLMODEL_H
