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
    void filterRequest(QString filter);
    void requestError(QString error);

public slots:
    void setFilter(QString filter);

    void updateFromServer();

protected:
    virtual void setFilterImpl(const QString &filter) = 0;
    virtual void updateFromServerImpl(quint32 leftAttempts) = 0;
    virtual int getLastRole() const = 0;

    virtual QSqlQuery &getQuery() = 0;
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

    void emitRequestError(const QString &err)
    { emit requestError(err); }

    /// reimplemented qt methods
    QVariant data(const QModelIndex &item, int role) const override;
    QHash<int, QByteArray> roleNames() const override;

private slots:
    void _setFilter(QString filter);

private:
    mutable QHash<int, QByteArray> mRoleNames;

    quint32 mRequestAttemptsCount;
    quint32 mRequestBatchSize;

    ServerApi *mApi;
    IcoImageProvider *mImageProvider;
    QSettings *mSettings;
};

#endif // LISTSQLMODEL_H
