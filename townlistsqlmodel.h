#ifndef TOWNLISTSQLMODEL_H
#define TOWNLISTSQLMODEL_H

#include <QtSql/QSqlQueryModel>
#include <QtSql/QSqlQuery>

class TownListSqlModel : public QSqlQueryModel
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

    int rowCount(const QModelIndex &parent = QModelIndex()) const override;

    QVariant data(const QModelIndex &item, int role) const override;

public slots:
    void setFilter(QString filterStr);

protected:
    QHash<int, QByteArray> roleNames() const override;

private:
    QHash<int, QByteArray> mRoleNames;
    QString mQueryMask;
    const QString mConnectionName;
    QSqlQuery mQuery;
};

#endif // TOWNLISTSQLMODEL_H
