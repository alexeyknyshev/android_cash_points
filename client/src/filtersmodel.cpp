#include "filtersmodel.h"

FiltersModel::FiltersModel(const QString &connectionName, ServerApi *api, IcoImageProvider *provider, QSettings *settings)
    : ListSqlModel(connectionName, api, provider, settings),
      mLastId(0)
{
    setRoleName(IdRole,        "filter_id");
    setRoleName(NameRole,      "filter_name");
    setRoleName(DataRole,      "filter_data");
    setRoleName(CurrentRole,   "filter_current");
    setRoleName(RemovableRole, "filter_removable");
}

FiltersModel::~FiltersModel()
{
}

QVariant FiltersModel::data(const QModelIndex &item, int role) const
{
    if (role == DataRole) {
        QVariant dynamic = item.data(DynamicRole);
        if (dynamic.isValid()) {
            DynamicFilter *filter = (DynamicFilter *)dynamic.value<void *>();
            return filter->createFilter();
        }
    }

    if (role < Qt::UserRole || role >= RoleLast)
    {
        return ListSqlModel::data(item, role);
    }

    return QStandardItemModel::data(index(item.row(), 0), role);
}

bool FiltersModel::setData(const QModelIndex &index,
                           const QVariant &value,
                           int role)
{
    if (role == CurrentRole) {
        const bool current = index.data(IdRole).toInt();
        if (!current) {
            const int rows = rowCount();
            for (int i = 0; i < rows; i++) {
                QStandardItem *it = item(i, index.column());
                it->setData(false, CurrentRole);
            }
            item(index.row(), index.column())->setData(true);
        }
        return true;
    } else {
        return ListSqlModel::setData(index, value, role);
    }
}

int FiltersModel::addFilter(QString filter)
{
    const int id = getNextId();
    QStandardItem *item = new QStandardItem;
    item->setData(id, IdRole);
    item->setData(false, CurrentRole);
    item->setData(filter, DataRole);

    insertRow(0, { item });

    return id;
}

/// TODO: fix memleak there
int FiltersModel::addDynamicFilter(DynamicFilter *filter)
{
    const int id = getNextId();
    QStandardItem *item = new QStandardItem;
    item->setData(id, IdRole);
    item->setData(false, CurrentRole);
    item->setData(qVariantFromValue((void *)filter), DynamicRole);

    insertRow(0, { item });

    return id;
}


bool FiltersModel::removeFilter(int id)
{
    const int rows = rowCount();
    for (int i = 0; i < rows; i++) {
        QModelIndex idx = index(i, 0);
        if (idx.data(IdRole).toInt() == id) {
            return removeRow(idx.row());
        }
    }
    return false;
}

QString FiltersModel::getCurrentFilter() const
{
    const int rows = rowCount();
    for (int i = 0; i < rows; i++) {
        QStandardItem *it = item(i, 0);
        if (it->data(CurrentRole).toBool()) {
            return it->data(DataRole).toString();
        }
    }
    return "";
}

void FiltersModel::setFilterName(int id, QString name)
{
    const int rows = rowCount();
    for (int i = 0; i < rows; i++) {
        QStandardItem *it = item(i, 0);
        if (it->data(IdRole).toInt() == id) {
            it->setData(name, NameRole);
            return;
        }
    }
}

//void FiltersModel::setCurrentFilter(int id)
//{
//}

