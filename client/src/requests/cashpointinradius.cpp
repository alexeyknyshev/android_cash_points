#include "cashpointinradius.h"

#include "../cashpointsqlmodel.h"
#include "../serverapi.h"

#include <QtCore/QDebug>
#include <QtCore/QJsonObject>
#include <QtCore/QJsonParseError>
#include <QtCore/QJsonArray>

CashPointInRadius::CashPointInRadius(CashPointSqlModel *model)
    : CashPointRequest(model)
{
    registerStepHandlers({
                             STEP_HANDLER(fetchIds),
                             STEP_HANDLER(fetchCashpoints)
                         });
}

void CashPointInRadius::fetchIds(ServerApiPtr api, quint32 leftAttempts)
{
    if (data.isEmpty()) {
        return;
    }

    mId = api->sendRequest("/nearby/cashpoints", data,
    [this, api, leftAttempts](ServerApi::RequestStatusCode reqCode, ServerApi::HttpStatusCode httpCode, const QByteArray &data) {
        if (!isRunning()) {
            if (isDisposing()) {
                deleteLater();
            }
            return;
        }
        const int step = 0;

        if (reqCode == ServerApi::RSC_Timeout) {
            emitUpdate(leftAttempts - 1, step);
            return;
        }

        const QString errText = trUtf8("Cannot receive list of nearby cashpoints");
        if (reqCode != ServerApi::RSC_Ok) {
            emitError(ServerApi::requestStatusCodeText(reqCode));
            emitStepFinished(*api, step, false, errText);
            return;
        }

        if (httpCode != ServerApi::HSC_Ok) {
            qWarning() << "Server request http code: " << httpCode;
            emitUpdate(leftAttempts - 1, step);
            return;
        }

        QJsonParseError err;
        const QJsonDocument json = QJsonDocument::fromJson(data, &err);
        if (err.error != QJsonParseError::NoError) {
            emitError("CashPointInRadius: server response json parse error: " + err.errorString());
            emitStepFinished(*api, step, false, errText);
            return;
        }

        if (!json.isArray()) {
            emitError("CashPointInRadius: response is not an array!");
            emitStepFinished(*api, step, false, errText);
            return;
        }

        mCashpointsToProcess.clear();
        mResponse = new CashPointResponse;
        mResponse->type = CashPointResponse::CashpointData;

        const QJsonArray arr = json.array();
        const auto end = arr.constEnd();
        for (auto it = arr.constBegin(); it != end; it++) {
            const QJsonValue &val = *it;
            const int id = val.toInt();
            if (id > 0) {
                mCashpointsToProcess.append(id);
//                qDebug() << id;
            }
        }

        // no cashpoints in radius => send response now
        if (mCashpointsToProcess.empty()) {
            emitResponseReady(true);
            emitStepFinished(*api, getStepHandlers().count() - 1, true, trUtf8("There is no nearby cashpoints"));
        }

        const int timestamp = QDateTime::currentDateTime().toMSecsSinceEpoch() / 1000;
        const int outdateTimestamp = timestamp - 300;

        const QMap<quint32, int> cachedCashpoints = getModel()->getCachedCashpoints();
        const auto cacheEnd = cachedCashpoints.end();

        auto it = mCashpointsToProcess.begin();
        while (it != mCashpointsToProcess.end()) {
            const quint32 id = *it;
            mResponse->addVisiableCashpoint(id);

            const auto cit = cachedCashpoints.find(id);
            if (cit != cacheEnd) {
                const int cachedTimestamp = cit.value();

                // cache is uptodate, so we do not need to fetch this id
                if (outdateTimestamp < cachedTimestamp) {
                    QJsonObject obj = getModel()->getCachedCashpointData(id);
                    if (!obj.isEmpty()) {
                        it = mCashpointsToProcess.erase(it);
                        mResponse->addCashPointData(obj);
                        continue;
                    }
                }
            }
            it++;
        }

        /// TODO: We have not to clear model but may be only remove invisiable points
//        getModel()->clear();
        emitStepFinished(*api, step, true, trUtf8("List of nearby cashpoints received"));
    });
}

void CashPointInRadius::fetchCashpoints(ServerApiPtr api, quint32 leftAttempts)
{
    const int step = 1;
    qDebug() << "fetchCashpoints";

    if (mCashpointsToProcess.empty()) {
        return;
    }

    const quint32 cashpointsToProcess = qMin(getModel()->getRequestBatchSize(), (quint32)mCashpointsToProcess.size());
    if (cashpointsToProcess == 0) {
        emitStepFinished(*api, step, true, trUtf8("Data of nearby cashpoints received"));
        return;
    }

    QList<quint32> cachedCashpoints = getModel()->getCachedCashpoints().keys();

    QJsonArray requestCashpoints;
    QJsonArray requestCachedCashpoints;

    for (quint32 i = 0; i < cashpointsToProcess; ++i) {
        const quint32 id = mCashpointsToProcess.front();
        mCashpointsToProcess.removeFirst();

        if (cachedCashpoints.contains(id)) {
            requestCachedCashpoints.append(QJsonValue((int)id));
        } else {
            requestCashpoints.append(QJsonValue((int)id));
        }
    }

    const QJsonObject json
    {
        { "cashpoints", requestCashpoints },
        { "cached", requestCachedCashpoints }
    };

    const bool lastRequest = mCashpointsToProcess.empty();
    const quint32 totalAttempts = getModel()->getAttemptsCount();

    /// Get cashpoints data from list
    mId = api->sendRequest("/cashpoints", json,
    [this, api, leftAttempts, step, lastRequest, totalAttempts]
    (ServerApi::RequestStatusCode reqCode, ServerApi::HttpStatusCode httpCode, const QByteArray &data) {
        if (!isRunning()) {
            if (isDisposing()) {
                deleteLater();
            }
            return;
        }

        const QString errText = trUtf8("Cannot receive data of nearby cashpoints");
        if (reqCode == ServerApi::RSC_Timeout) {
            emitUpdate(leftAttempts - 1, step);
            return;
        }

        if (reqCode != ServerApi::RSC_Ok) {
            emitError(ServerApi::requestStatusCodeText(reqCode));
            emitStepFinished(*api, step, false, errText);
            return;
        }

        if (httpCode != ServerApi::HSC_Ok) {
            qWarning() << "Server request http code: " << httpCode;
            emitUpdate(leftAttempts - 1, step);
            return;
        }

        QJsonParseError err;
        const QJsonDocument json = QJsonDocument::fromJson(data, &err);
        if (err.error != QJsonParseError::NoError) {
            emitError("CashPointInRadius: server response json parse error: " + err.errorString());
            emitStepFinished(*api, step, false, errText);
            return;
        }

        if (!json.isArray()) {
            emitError("CashPointInRadius: server response json array expected. Dump: " +
                      QString::fromUtf8(json.toJson(QJsonDocument::Compact)));
            emitStepFinished(*api, step, false, errText);
            return;
        }
        const QJsonArray cashpoints = json.array();

        if (!mResponse) {
            mResponse = new CashPointResponse;
        }

        const auto cpEnd = cashpoints.end();
        for (auto it = cashpoints.begin(); it != cpEnd; it++) {
            mResponse->addCashPointData(it->toObject());
            //qDebug() << "added cashpoint: " << it->toObject()["id"].toInt();
        }

        emitResponseReady(lastRequest);
        if (lastRequest) {
            emitStepFinished(*api, step, true, trUtf8("Data of nearby cashpoints received"));
        } else {
            emitUpdate(totalAttempts, step);
        }
    });
}

bool CashPointInRadius::fromJson(const QJsonObject &json)
{
    const QJsonValue radius =    json["radius"];
    const QJsonValue latitude =  json["latitude"];
    const QJsonValue longitude = json["longitude"];
    const QJsonValue filter    = json["filter"];

    const QJsonValue topLeft =   json["topLeft"];
    const QJsonValue botRight =  json["bottomRight"];

    if (!radius.isDouble() || !latitude.isDouble() || !longitude.isDouble() ||
        !(filter.isObject() || filter.isUndefined()))
    {
        return false;
    }

    if (!topLeft.isObject() || !botRight.isObject()) {
        return false;
    }

    data["longitude"] = longitude;
    data["latitude"] = latitude;
    data["radius"] = radius;
    data["topLeft"] = topLeft;
    data["bottomRight"] = botRight;

    if (filter.isObject()) {
        data["filter"] = filter;
    }

    return true;
}
