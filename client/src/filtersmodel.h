#ifndef FILTERSMODEL_H
#define FILTERSMODEL_H

#include "listsqlmodel.h"

class FiltersModel : public ListSqlModel
{
    Q_OBJECT

public:
    enum Roles {
        IdRole = Qt::UserRole,
        NameRole,
        DataRole,
        CurrentRole,
        RemovableRole,
        DynamicRole,

        RoleLast
    };

    FiltersModel(const QString &connectionName, ServerApi *api,
                 IcoImageProvider *provider, QSettings *settings);
    ~FiltersModel();

    QVariant data(const QModelIndex &item, int role) const override;

    bool setData(const QModelIndex &index,
                 const QVariant &value,
                 int role) override;

    Q_INVOKABLE int addFilter(QString filter);
    Q_INVOKABLE bool removeFilter(int id);
    Q_INVOKABLE QString getCurrentFilter() const;

    class DynamicFilter {
    public:
        virtual QString createFilter() = 0;
    };

    int addDynamicFilter(DynamicFilter *filter);

public slots:
    void setFilterName(int id, QString name);
//    void setCurrentFilter(int id);

protected:
    int getNextId() {
        mLastId++;
        return mLastId;
    }

    int getLastRole() const override { return RoleLast; }
    virtual QSqlQuery *getQuery() override { return nullptr; }
    virtual bool needEscapeFilter() const override { return false; }
    void setFilterImpl(const QString &, const QJsonObject &) override {}
    void updateFromServerImpl(quint32) override {}
    QList<int> getSelectedIdsImpl() const override { return {}; }


private:
    int mLastId;
};

#endif // FILTERSMODEL_H
