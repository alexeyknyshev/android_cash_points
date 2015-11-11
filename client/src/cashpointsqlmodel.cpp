#include "cashpointsqlmodel.h"

#include <QtCore/QDebug>
#include <QtSql/QSqlRecord>
#include <QtPositioning/QGeoCoordinate>
#include <QtCore/QJsonDocument>
#include <QtCore/QJsonArray>

#include "rpctype.h"
#include "serverapi.h"

// CashPoint request types
#define RT_RADUS "radius"
#define RT_TOWN "town"

struct CashPoint : public RpcType<CashPoint>
{
    QString type;
    quint32 bankId;
    quint32 townId;
    qreal longitude;
    qreal latitude;
    QString address;
    QString addressComment;
    QString metroName;
    bool mainOffice;
    bool withoutWeekend;
    bool roundTheClock;
    bool worksAsShop;
    bool rub;
    bool usd;
    bool eur;
    bool cashIn;

    CashPoint()
        : bankId(0),
          townId(0),
          longitude(0.0f),
          latitude(0.0f),
          mainOffice(false),
          withoutWeekend(false),
          roundTheClock(false),
          worksAsShop(false),
          rub(false),
          usd(false),
          eur(false),
          cashIn(false)
    { }

    static CashPoint fromJsonObject(const QJsonObject &obj)
    {
        CashPoint result = RpcType<CashPoint>::fromJsonObject(obj);

        result.type           = obj["type"].toString();
        result.bankId         = obj["bank_id"].toInt();
        result.townId         = obj["town_id"].toInt();
        result.longitude      = obj["longitude"].toDouble();
        result.latitude       = obj["latitude"].toDouble();
        result.address        = obj["address"].toString();
        result.addressComment = obj["address_comment"].toString();
        result.metroName      = obj["metro_name"].toString();
        result.mainOffice     = obj["main_office"].isBool();
        result.withoutWeekend = obj["without_weekend"].toBool();
        result.roundTheClock  = obj["round_the_clock"].toBool();
        result.worksAsShop    = obj["works_as_shop"].toBool();
        result.rub            = obj["rub"].toBool();
        result.usd            = obj["usd"].toBool();
        result.eur            = obj["eur"].toBool();

        return result;
    }
};

/// ================================================

class CashPointRequest : public QObject
{
    Q_OBJECT

public:
    CashPointRequest(CashPointSqlModel *model)
        : mModel(model)
    {
        connect(this, SIGNAL(update(quint32)), SLOT(send(quint32)), Qt::QueuedConnection);
        connect(this, SIGNAL(error(QString)), mModel, SIGNAL(requestError(QString)));
    }

    virtual void sendImpl(ServerApi *api, quint32 leftAttepmts) = 0;
    const QDateTime &getLastUpdateTime() const { return mLastUpdateTime; }

signals:
    void update(quint32 leftAttempts);
    void error(QString err);

public slots:
    void send(quint32 leftAttempts) {
        sendImpl(getModel()->getServerApi(), leftAttempts);
    }

protected:
    CashPointSqlModel *getModel() const { return mModel; }

    void emitUpdate(quint32 leftAttempts) {
        emit update(leftAttempts);
    }

    void emitError(QString err) {
        emit error(err);
    }

    void setLastUpdateTime(const QDateTime &time) {
        mLastUpdateTime = time;
    }

private:
    CashPointSqlModel *const mModel;
    QDateTime mLastUpdateTime;
};

// =================================================

class CashPointInRadius : public CashPointRequest
{
public:
    CashPointInRadius(CashPointSqlModel *model)
        : CashPointRequest(model),
          mRadius(1000.0f)
    { }

    void sendImpl(ServerApi *api, quint32 leftAttempts) override {
        if (leftAttempts == 0) {
            qDebug() << "CashPointInRadius: no retry attempt left";
            return;
        }

        QJsonObject json;
        json["longitude"] = mCoord.longitude();
        json["latitude"] = mCoord.latitude();
        json["radius"] = mRadius;
        api->sendRequest("/nearby/cashpoints", json,
        [&](ServerApi::RequestStatusCode reqCode, ServerApi::HttpStatusCode httpCode, const QByteArray &data) {
            if (reqCode == ServerApi::RSC_Timeout) {
                emitUpdate(leftAttempts - 1);
                return;
            }

            if (reqCode != ServerApi::RSC_Ok) {
                emitError(ServerApi::requestStatusCodeText(reqCode));
                return;
            }

            if (httpCode != ServerApi::HSC_Ok) {
                qWarning() << "Server request http code: " << httpCode;
                emitUpdate(leftAttempts - 1);
                return;
            }

            QJsonParseError err;
            const QJsonDocument json = QJsonDocument::fromJson(data, &err);
            if (err.error != QJsonParseError::NoError) {
                emitError("CashPointInRadius: server response json parse error: " + err.errorString());
                return;
            }

            setLastUpdateTime(QDateTime::currentDateTime());
        });
    }

    void setRadius(qreal radius) {
        if (radius <= 0) {
            qDebug() << "cashpoint search radius must be positive";
            return;
        }
        mRadius = radius;
    }

    void setCoordinate(const QGeoCoordinate &coord) {
        if (!coord.isValid()) {
            qDebug() << "cashpoinst search coordinate must be valid";
            return;
        }
        mCoord = coord;
    }

private:
    qreal mRadius;
    QGeoCoordinate mCoord;
};

/// ================================================

CashPointSqlModel::CashPointSqlModel(const QString &connectionName,
                                     ServerApi *api,
                                     IcoImageProvider *imageProvider,
                                     QSettings *settings)
    : ListSqlModel(connectionName, api, imageProvider, settings),
      mQuery(QSqlDatabase::database(connectionName)),
      mRequest(nullptr)
{
    setRoleName(IdRole,             "cp_id");
    setRoleName(TypeRole,           "cp_type");
    setRoleName(BankIdRole,         "cp_bank_id");
    setRoleName(TownIdRole,         "cp_town_id");
    setRoleName(LongitudeRole,      "cp_coord_lon");
    setRoleName(LatitudeRole,       "cp_coord_lat");
    setRoleName(AddressRole,        "cp_address");
    setRoleName(AddressCommentRole, "cp_address_comment");
    setRoleName(MetroNameRole,      "cp_metro_name");
    setRoleName(MainOfficeRole,     "cp_main_office");
    setRoleName(WithoutWeekendRole, "cp_without_weekend");
    setRoleName(RoundTheClockRole,  "cp_round_the_clock");
    setRoleName(WorksAsShopRole,    "cp_works_as_shop");
    setRoleName(RubRole,            "cp_rub");
    setRoleName(UsdRole,            "cp_usd");
    setRoleName(EurRole,            "cp_eur");
    setRoleName(CashInRole,         "cp_cash_in");

    connect(this, SIGNAL(delayedUpdate()), SLOT(updateFromServer()), Qt::QueuedConnection);
}

QVariant CashPointSqlModel::data(const QModelIndex &item, int role) const
{

}

bool CashPointSqlModel::setData(const QModelIndex &index, const QVariant &value, int role)
{

}

void CashPointSqlModel::sendRequest(CashPointRequest *request)
{
    if (mRequest == request) {
        return;
    }

    if (mRequest) {
        mRequest->deleteLater();
    }

    mRequest = request;
    if (request) {
        request->send(getAttemptsCount());
    }

    emit delayedUpdate();
}

void CashPointSqlModel::updateFromServerImpl(quint32 leftAttempts)
{
    if (leftAttempts == 0) {
        return;
    }


}

void CashPointSqlModel::setFilterImpl(const QString &filter)
{
    if (filter.isEmpty()) {
        return;
    }

    QJsonParseError err;
    const QJsonDocument json = QJsonDocument::fromJson(filter.toUtf8(), &err);
    if (err.error != QJsonParseError::NoError) {
        setFilterFreeForm(filter);
        return;
    }

    if (!json.isObject()) {
        emitRequestError("CashPointSqlModel::setFilterImpl: Cannot local request is not a json object.");
        return;
    }

    setFilterJson(json.object());
}

void CashPointSqlModel::setFilterJson(const QJsonObject &json)
{
    const QString type = json["type"].toString();

    CashPointRequest *req = nullptr;
    if (type == RT_RADUS) {
        CashPointInRadius *tmpReq = new CashPointInRadius(this);
        tmpReq->setRadius(json["radius"].toDouble());
        QGeoCoordinate coord;
        coord.setLatitude(json["latitude"].toDouble());
        coord.setLongitude(json["longitude"].toDouble());
        tmpReq->setCoordinate(coord);
        req = tmpReq;
    } else {
        emitRequestError("ashPointSqlModel::setFilterJson: unknown req type: " + type);
        return;
    }

    sendRequest(req);
}

void CashPointSqlModel::setFilterFreeForm(const QString &filter)
{

}
