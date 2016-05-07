#ifndef LISTSQLMODEL_H
#define LISTSQLMODEL_H

#include <QtGui/QStandardItemModel>

class ServerApi;
class IcoImageProvider;
class QSqlQuery;
class QSettings;

class ListSqlModel : public QStandardItemModel
{
    Q_OBJECT

public:
    ListSqlModel(const QString &connectionName,
                 ServerApi *api,
                 IcoImageProvider *imageProvider,
                 QSettings *settings);

    static QString escapeFilter(QString filter);

    ServerApi *getServerApi() const { return mApi; }
    quint32 getAttemptsCount() const { return mRequestAttemptsCount; }
    quint32 getRequestBatchSize() const { return mRequestBatchSize; }

signals:
    void serverDataReceived();
    void requestError(int id, QString msg);
    void filterRequest(QString filter, QString options);

    void updateProgress(int done, int total);

public slots:
    void setFilter(QString filter, QString options);

    void updateFromServer();

protected:
    ListSqlModel(ListSqlModel *submodel);
    virtual void setFilterImpl(const QString &filter, const QJsonObject &options) = 0;
    virtual void updateFromServerImpl(quint32 leftAttempts) = 0;
    virtual int getLastRole() const = 0;

    virtual QSqlQuery *getQuery() = 0;
    virtual bool needEscapeFilter() const = 0;

    void setAttemptsCount(quint32 count) { mRequestAttemptsCount = count; }

    void setRequestBatchSize(quint32 size) {
        Q_ASSERT_X(size > 0, "setRequestBatchSize", "zero batch size");
        mRequestBatchSize = size;
    }

    IcoImageProvider *getIcoImageProvider() const { return mImageProvider; }
    QSettings *getSettings() const { return mSettings; }

    void setRoleName(int role, const QByteArray &name) const { mRoleNames[role] = name; }

    void emitServerDataReceived()
    { emit serverDataReceived(); }

    void emitRequestError(int requestId, const QString &msg)
    { emit requestError(requestId, msg); }

    void emitUpdateProgress()
    { emit updateProgress(mUploadedCount, mExpectedUploadCount); }

    void setUploadedCount(int count)
    { mUploadedCount = count; }

    int getUploadedCount() const
    { return mUploadedCount; }

    void setExpectedUploadCount(int count)
    { mExpectedUploadCount = count; }

    int getExpectedUploadCount() const
    { return mExpectedUploadCount; }

    /// reimplemented qt methods
    QVariant data(const QModelIndex &item, int role) const override;
    QHash<int, QByteArray> roleNames() const override;

    const QString &getDBConnectionName() const
    { return mDBConnectionName; }

private slots:
    void _setFilter(QString filter, QString options);

private:
    mutable QHash<int, QByteArray> mRoleNames;

    quint32 mRequestAttemptsCount;
    quint32 mRequestBatchSize;

    ServerApi *mApi;
    IcoImageProvider *mImageProvider;
    QSettings *mSettings;

    const QString mDBConnectionName;

    int mExpectedUploadCount;
    int mUploadedCount;
};

#endif // LISTSQLMODEL_H
