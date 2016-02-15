#include "emptyimageprovider.h"

#include <QtGui/QPainter>

EmptyImageProvider::EmptyImageProvider()
    : QQuickImageProvider(QQmlImageProviderBase::Image)
{ }

QImage EmptyImageProvider::requestImage(const QString &, QSize *size, const QSize &requestedSize)
{
    if (size) {
        *size = requestedSize;
    }

    QImage image(requestedSize, QImage::Format_ARGB32);
    image.fill(Qt::transparent);
    return image;
}

