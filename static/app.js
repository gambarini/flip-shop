// Global state
let cart = null;
let items = [];
let itemQuantities = {}; // Track quantities user wants to add for each item

// API base URL (adjust if needed)
const API_BASE = '';

// Helper: Format cents to dollar string
function formatPrice(cents) {
    return `$${(cents / 100).toFixed(2)}`;
}

// Helper: Show alert message
function showAlert(message, type = 'success') {
    const alertEl = document.getElementById('alert');
    alertEl.textContent = message;
    alertEl.className = `alert ${type}`;
    alertEl.style.display = 'block';

    setTimeout(() => {
        alertEl.style.display = 'none';
    }, 5000);
}

// Helper: Show error
function showError(message) {
    showAlert(message, 'error');
}

// API: Fetch all items
async function fetchItems() {
    try {
        const response = await fetch(`${API_BASE}/items`);
        if (!response.ok) throw new Error('Failed to fetch items');
        const data = await response.json();
        items = data || [];
        renderItems();
    } catch (error) {
        showError(`Error loading items: ${error.message}`);
        console.error('Fetch items error:', error);
    }
}

// API: Create a new cart
async function createCart() {
    try {
        const response = await fetch(`${API_BASE}/cart`, {
            method: 'POST'
        });
        if (!response.ok) throw new Error('Failed to create cart');
        cart = await response.json();
        updateCartDisplay();
    } catch (error) {
        showError(`Error creating cart: ${error.message}`);
        console.error('Create cart error:', error);
    }
}

// API: Add item to cart
async function addToCart(sku, qty) {
    if (!cart) return;

    try {
        const response = await fetch(`${API_BASE}/cart/${cart.CartID}/purchase`, {
            method: 'PUT',
            headers: {
                'Content-Type': 'application/json'
            },
            body: JSON.stringify({ sku, qty })
        });

        if (!response.ok) {
            const error = await response.json();
            throw new Error(error.error || 'Failed to add item');
        }

        cart = await response.json();
        updateCartDisplay();
        await fetchItems(); // Refresh to show updated availability
        showAlert(`Added ${qty}x item to cart`);

        // Reset quantity input for this item
        itemQuantities[sku] = 1;
        renderItems();
    } catch (error) {
        showError(`Error adding to cart: ${error.message}`);
        console.error('Add to cart error:', error);
    }
}

// API: Remove item from cart
async function removeFromCart(sku) {
    if (!cart) return;

    const purchase = cart.Purchases[sku];
    if (!purchase) return;

    try {
        const response = await fetch(`${API_BASE}/cart/${cart.CartID}/purchase`, {
            method: 'DELETE',
            headers: {
                'Content-Type': 'application/json'
            },
            body: JSON.stringify({ sku, qty: purchase.Qty })
        });

        if (!response.ok) {
            const error = await response.json();
            throw new Error(error.error || 'Failed to remove item');
        }

        cart = await response.json();
        updateCartDisplay();
        await fetchItems(); // Refresh to show updated availability
        showAlert('Item removed from cart');
    } catch (error) {
        showError(`Error removing from cart: ${error.message}`);
        console.error('Remove from cart error:', error);
    }
}

// API: Submit cart
async function submitCart() {
    if (!cart) return;

    try {
        const response = await fetch(`${API_BASE}/cart/${cart.CartID}/status/submitted`, {
            method: 'PUT'
        });

        if (!response.ok) {
            const error = await response.json();
            throw new Error(error.error || 'Failed to submit cart');
        }

        cart = await response.json();
        updateCartDisplay();
        await fetchItems(); // Refresh to show updated availability
        showAlert('Cart submitted! Promotions applied successfully.', 'success');
    } catch (error) {
        showError(`Error submitting cart: ${error.message}`);
        console.error('Submit cart error:', error);
    }
}

// Render items grid
function renderItems() {
    const grid = document.getElementById('itemsGrid');

    if (items.length === 0) {
        grid.innerHTML = '<div class="loading">No items available</div>';
        return;
    }

    grid.innerHTML = items.map(item => {
        const available = item.QtyAvailable - item.QtyReserved;
        const isOutOfStock = available <= 0;
        const isLowStock = available > 0 && available <= 2;
        const qty = itemQuantities[item.Sku] || 1;
        const isCartSubmitted = cart && cart.CartStatus === 'Submitted';

        let availabilityClass = '';
        let availabilityText = `${available} available`;

        if (isOutOfStock) {
            availabilityClass = 'out-of-stock';
            availabilityText = 'Out of stock';
        } else if (isLowStock) {
            availabilityClass = 'low-stock';
            availabilityText = `Only ${available} left!`;
        }

        return `
            <div class="item-card">
                <h3>${item.Name}</h3>
                <div class="item-sku">SKU: ${item.Sku}</div>
                <div class="item-price">${formatPrice(item.Price)}</div>
                <div class="item-availability ${availabilityClass}">${availabilityText}</div>

                <div class="quantity-control">
                    <button onclick="decrementQty('${item.Sku}')" ${qty <= 1 ? 'disabled' : ''}>-</button>
                    <input type="number" value="${qty}" min="1" max="${available}"
                           onchange="setQty('${item.Sku}', this.value)"
                           ${isOutOfStock || isCartSubmitted ? 'disabled' : ''}>
                    <button onclick="incrementQty('${item.Sku}')" ${qty >= available ? 'disabled' : ''}>+</button>
                </div>

                <button class="btn btn-primary"
                        onclick="addToCart('${item.Sku}', ${qty})"
                        ${isOutOfStock || isCartSubmitted ? 'disabled' : ''}>
                    ${isCartSubmitted ? 'Cart Submitted' : 'Add to Cart'}
                </button>
            </div>
        `;
    }).join('');
}

// Quantity control functions
function incrementQty(sku) {
    const item = items.find(i => i.Sku === sku);
    if (!item) return;

    const available = item.QtyAvailable - item.QtyReserved;
    const current = itemQuantities[sku] || 1;

    if (current < available) {
        itemQuantities[sku] = current + 1;
        renderItems();
    }
}

function decrementQty(sku) {
    const current = itemQuantities[sku] || 1;
    if (current > 1) {
        itemQuantities[sku] = current - 1;
        renderItems();
    }
}

function setQty(sku, value) {
    const item = items.find(i => i.Sku === sku);
    if (!item) return;

    const available = item.QtyAvailable - item.QtyReserved;
    const qty = parseInt(value) || 1;

    itemQuantities[sku] = Math.max(1, Math.min(qty, available));
    renderItems();
}

// Update cart display
function updateCartDisplay() {
    if (!cart) return;

    // Update cart ID and status
    document.getElementById('cartId').textContent = cart.CartID.substring(0, 8) + '...';

    const statusBadge = document.querySelector('.status-badge');
    statusBadge.textContent = cart.CartStatus;
    statusBadge.className = `status-badge ${cart.CartStatus.toLowerCase()}`;

    // Update cart items
    const cartItemsEl = document.getElementById('cartItems');
    const purchases = cart.Purchases || {};
    const purchaseKeys = Object.keys(purchases);

    if (purchaseKeys.length === 0) {
        cartItemsEl.innerHTML = '<p class="empty-cart">Your cart is empty. Add items to get started!</p>';
        document.getElementById('cartSummary').style.display = 'none';
        document.getElementById('submitBtn').disabled = true;
        return;
    }

    // Render cart items
    const isSubmitted = cart.CartStatus === 'Submitted';
    cartItemsEl.innerHTML = purchaseKeys.map(sku => {
        const purchase = purchases[sku];
        const lineTotal = purchase.Price * purchase.Qty;
        const finalPrice = lineTotal - purchase.Discount;

        return `
            <div class="cart-item ${isSubmitted ? 'submitted' : ''}">
                <div class="cart-item-info">
                    <h4>${purchase.Name}</h4>
                    <div class="cart-item-details">
                        ${formatPrice(purchase.Price)} × ${purchase.Qty}
                    </div>
                    ${!isSubmitted ? `
                        <div class="cart-item-actions">
                            <button class="btn btn-remove" onclick="removeFromCart('${sku}')">
                                Remove
                            </button>
                        </div>
                    ` : ''}
                </div>
                <div class="cart-item-price">
                    <div class="price">${formatPrice(finalPrice)}</div>
                    ${purchase.Discount > 0 ? `
                        <div class="discount">-${formatPrice(purchase.Discount)} discount</div>
                    ` : ''}
                </div>
            </div>
        `;
    }).join('');

    // Update summary
    document.getElementById('cartSummary').style.display = 'block';

    let subtotal = 0;
    let totalDiscount = 0;

    purchaseKeys.forEach(sku => {
        const purchase = purchases[sku];
        subtotal += purchase.Price * purchase.Qty;
        totalDiscount += purchase.Discount;
    });

    document.getElementById('subtotal').textContent = formatPrice(subtotal);

    if (totalDiscount > 0) {
        document.getElementById('discountRow').style.display = 'flex';
        document.getElementById('totalDiscount').textContent = `-${formatPrice(totalDiscount)}`;
    } else {
        document.getElementById('discountRow').style.display = 'none';
    }

    const total = isSubmitted ? cart.Total : subtotal;
    document.getElementById('total').textContent = formatPrice(total);

    // Update submit button
    const submitBtn = document.getElementById('submitBtn');
    if (isSubmitted) {
        submitBtn.textContent = 'Cart Submitted ✓';
        submitBtn.disabled = true;
        submitBtn.className = 'btn btn-success btn-submit';
    } else {
        submitBtn.textContent = 'Submit Cart & Apply Promotions';
        submitBtn.disabled = false;
        submitBtn.className = 'btn btn-primary btn-submit';
    }
}

// Event listeners
document.getElementById('submitBtn').addEventListener('click', submitCart);

// Initialize app
async function init() {
    await createCart();
    await fetchItems();
}

// Start the app
init();
