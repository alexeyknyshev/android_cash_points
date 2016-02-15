#include "listsqlmodel.h"

#define DEFAULT_ATTEMPTS_COUNT 3
#define DEFAULT_BATCH_SIZE 128

ListSqlModel::ListSqlModel(const QString &connectionName,
                           ServerApi *api,
                           IcoImageProvider *imageProvider,
                           QSettings *settings)
    : mApi(api),
      mImageProvider(imageProvider),
      mSettings(settings),
      mDBConnectionName(connectionName)
{
    Q_ASSERT_X(api, "ListSqlModel(const QString &, ServerApi *, IcoImageProvider *, QSettings *)", "null ServerApi ptr");
    Q_ASSERT_X(imageProvider, "ListSqlModel(const QString &, ServerApi *, IcoImageProvider *, QSettings *)", "null IcoImageProvider ptr");
    Q_ASSERT_X(settings, "ListSqlModel(const QString &, ServerApi *, IcoImageProvider *, QSettings *)", "null QSettiings ptr");

    setAttemptsCount(DEFAULT_ATTEMPTS_COUNT);
    setRequestBatchSize(DEFAULT_BATCH_SIZE);

    connect(this, SIGNAL(filterRequest(QString)),
            this, SLOT(_setFilter(QString)), Qt::QueuedConnection);
}

ListSqlModel::ListSqlModel(ListSqlModel *submodel)
    : mDBConnectionName(submodel->getDBConnectionName())
{
    Q_ASSERT_X(submodel, "ListSqlModel(ListSqlModel *)", "null ListSqlModel ptr");

    mApi = submodel->getServerApi();
    mImageProvider = submodel->getIcoImageProvider();
    mSettings = submodel->getSettings();

    setAttemptsCount(DEFAULT_ATTEMPTS_COUNT);
    setRequestBatchSize(DEFAULT_BATCH_SIZE);

    connect(this, SIGNAL(filterRequest(QString)),
            this, SLOT(_setFilter(QString)), Qt::QueuedConnection);
}

QString ListSqlModel::escapeFilter(QString filter)
{
    if (filter.isEmpty()) {
//        return "";
    }

    filter.replace('_', "");
    filter.replace('%', "");
    filter.replace('*', '%');
    filter.replace('?', '_');

    if (!filter.startsWith('%'))
    {
        filter.prepend('%');
    }

    if (!filter.endsWith('%'))
    {
        filter.append('%');
    }

    return filter;
}

void ListSqlModel::setFilter(QString filter)
{
    if (needEscapeFilter()) {
        emit filterRequest(escapeFilter(filter));
    } else {
        emit filterRequest(filter);
    }
}

void ListSqlModel::_setFilter(QString filter)
{
    setFilterImpl(filter);
}

void ListSqlModel::updateFromServer()
{
    updateFromServerImpl(getAttemptsCount());
}

QVariant ListSqlModel::data(const QModelIndex &item, int role) const
{
    if (role == getLastRole()) {
        return item.row();
    }

    return QStandardItemModel::data(item, role);
}

QHash<int, QByteArray> ListSqlModel::roleNames() const
{
    int lastRole = getLastRole();
    if (!mRoleNames.contains(lastRole)) {
        setRoleName(lastRole, "index");
    }

    return mRoleNames;
}
