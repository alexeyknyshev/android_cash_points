#ifndef RPCTYPE_H
#define RPCTYPE_H

#include <QJsonObject>
#include <QJsonArray>

class QStandardItem;

template<typename T>
struct RpcType
{
    RpcType()
        : id(0)
    { }

    virtual ~RpcType() { }

    quint32 id;

    bool isValid() const
    {
        return id != 0;
    }

    static T fromJsonObject(const QJsonObject &obj)
    {
        T result;
        result.id = obj["id"].toInt();
        return result;
    }

    static QList<T> fromJsonArray(const QJsonArray &arr) {
        QList<T> result;
        for (const QJsonValue &val : arr) {
            Q_ASSERT(val.isObject());
            const T t = T::fromJsonObject(val.toObject());
            if (t.isValid()) {
                result.append(t);
            }
        }
        return result;
    }

    virtual void fillItem(QStandardItem *out) const = 0;
};

#endif // RPCTYPE_H
